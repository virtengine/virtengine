package app

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/sha3"
)

// NoiseProtocolName is the protocol name for our Noise implementation
// NK pattern: No static key for initiator, Known static key for responder
const NoiseProtocolName = "Noise_NK_25519_ChaChaPoly_SHA256"

// NoisePatternNK implements the NK handshake pattern:
// <- s
// ...
// -> e, es
// <- e, ee

// NoiseHandshakeState represents the state during a Noise Protocol handshake.
type NoiseHandshakeState struct {
	// Local keys
	localStatic    *NoiseKeyPair
	localEphemeral *NoiseKeyPair

	// Remote keys
	remoteStatic    []byte
	remoteEphemeral []byte

	// Symmetric state
	chainingKey []byte
	handshakeHash []byte

	// Direction
	isInitiator bool

	// Configuration
	config NoiseConfig

	mu sync.RWMutex
}

// NoiseKeyPair holds a Curve25519 key pair.
type NoiseKeyPair struct {
	PrivateKey []byte
	PublicKey  []byte
}

// NoiseSession represents an established encrypted session.
type NoiseSession struct {
	// Cipher states for sending and receiving
	sendCipher   cipher.AEAD
	recvCipher   cipher.AEAD
	sendNonce    uint64
	recvNonce    uint64
	
	// Session identification
	remotePublicKey []byte
	handshakeHash   []byte

	// Underlying connection
	conn net.Conn

	// Read/write buffers
	readBuf  []byte
	writeBuf []byte

	mu sync.Mutex
}

// NoiseTransport wraps a connection with Noise Protocol encryption.
type NoiseTransport struct {
	config        NoiseConfig
	localKeyPair  *NoiseKeyPair
	trustedKeys   map[string]bool // Map of trusted remote public keys (hex encoded)
	
	mu sync.RWMutex
}

// GenerateNoiseKeyPair generates a new Curve25519 key pair.
func GenerateNoiseKeyPair() (*NoiseKeyPair, error) {
	var privateKey [32]byte
	if _, err := io.ReadFull(rand.Reader, privateKey[:]); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}

	// Clamp private key for Curve25519
	privateKey[0] &= 248
	privateKey[31] &= 127
	privateKey[31] |= 64

	var publicKey [32]byte
	curve25519.ScalarBaseMult(&publicKey, &privateKey)

	return &NoiseKeyPair{
		PrivateKey: privateKey[:],
		PublicKey:  publicKey[:],
	}, nil
}

// NewNoiseTransport creates a new Noise Protocol transport.
func NewNoiseTransport(config NoiseConfig) (*NoiseTransport, error) {
	keyPair, err := GenerateNoiseKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	return &NoiseTransport{
		config:       config,
		localKeyPair: keyPair,
		trustedKeys:  make(map[string]bool),
	}, nil
}

// GetPublicKey returns the local static public key.
func (t *NoiseTransport) GetPublicKey() []byte {
	return t.localKeyPair.PublicKey
}

// AddTrustedKey adds a trusted remote public key.
func (t *NoiseTransport) AddTrustedKey(publicKey []byte) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.trustedKeys[fmt.Sprintf("%x", publicKey)] = true
}

// RemoveTrustedKey removes a trusted remote public key.
func (t *NoiseTransport) RemoveTrustedKey(publicKey []byte) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.trustedKeys, fmt.Sprintf("%x", publicKey))
}

// IsTrusted checks if a remote public key is trusted.
func (t *NoiseTransport) IsTrusted(publicKey []byte) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.trustedKeys[fmt.Sprintf("%x", publicKey)]
}

// SecureOutbound performs a Noise Protocol handshake as initiator.
func (t *NoiseTransport) SecureOutbound(conn net.Conn, remoteStaticKey []byte) (*NoiseSession, error) {
	if len(remoteStaticKey) != 32 {
		return nil, errors.New("invalid remote static key length")
	}

	// Set handshake deadline
	if t.config.HandshakeTimeout > 0 {
		if err := conn.SetDeadline(time.Now().Add(t.config.HandshakeTimeout)); err != nil {
			return nil, fmt.Errorf("failed to set handshake deadline: %w", err)
		}
	}

	// Initialize handshake state
	state, err := t.newHandshakeState(true, remoteStaticKey)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize handshake: %w", err)
	}

	// Generate ephemeral key pair
	state.localEphemeral, err = GenerateNoiseKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ephemeral key: %w", err)
	}

	// -> e, es
	// Mix ephemeral public key
	state.mixHash(state.localEphemeral.PublicKey)
	
	// Perform DH: es
	sharedSecret, err := curve25519DH(state.localEphemeral.PrivateKey, remoteStaticKey)
	if err != nil {
		return nil, fmt.Errorf("DH es failed: %w", err)
	}
	state.mixKey(sharedSecret)

	// Send message 1: e
	msg1 := make([]byte, 32)
	copy(msg1, state.localEphemeral.PublicKey)
	if _, err := conn.Write(msg1); err != nil {
		return nil, fmt.Errorf("failed to send message 1: %w", err)
	}

	// <- e, ee
	// Receive remote ephemeral
	msg2 := make([]byte, 48) // 32 bytes ephemeral + 16 bytes auth tag
	if _, err := io.ReadFull(conn, msg2); err != nil {
		return nil, fmt.Errorf("failed to receive message 2: %w", err)
	}

	state.remoteEphemeral = msg2[:32]
	state.mixHash(state.remoteEphemeral)

	// Perform DH: ee
	sharedSecret, err = curve25519DH(state.localEphemeral.PrivateKey, state.remoteEphemeral)
	if err != nil {
		return nil, fmt.Errorf("DH ee failed: %w", err)
	}
	state.mixKey(sharedSecret)

	// Verify authentication tag
	if err := state.verifyTag(msg2[32:]); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Split to transport keys
	session, err := state.split(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to split to transport: %w", err)
	}
	session.remotePublicKey = remoteStaticKey

	// Clear deadline
	if err := conn.SetDeadline(time.Time{}); err != nil {
		return nil, fmt.Errorf("failed to clear deadline: %w", err)
	}

	return session, nil
}

// SecureInbound performs a Noise Protocol handshake as responder.
func (t *NoiseTransport) SecureInbound(conn net.Conn) (*NoiseSession, error) {
	// Set handshake deadline
	if t.config.HandshakeTimeout > 0 {
		if err := conn.SetDeadline(time.Now().Add(t.config.HandshakeTimeout)); err != nil {
			return nil, fmt.Errorf("failed to set handshake deadline: %w", err)
		}
	}

	// Initialize handshake state
	state, err := t.newHandshakeState(false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize handshake: %w", err)
	}
	state.localStatic = t.localKeyPair

	// <- e, es
	// Receive remote ephemeral
	msg1 := make([]byte, 32)
	if _, err := io.ReadFull(conn, msg1); err != nil {
		return nil, fmt.Errorf("failed to receive message 1: %w", err)
	}

	state.remoteEphemeral = msg1
	state.mixHash(state.remoteEphemeral)

	// Perform DH: es (from responder perspective: se)
	sharedSecret, err := curve25519DH(t.localKeyPair.PrivateKey, state.remoteEphemeral)
	if err != nil {
		return nil, fmt.Errorf("DH se failed: %w", err)
	}
	state.mixKey(sharedSecret)

	// -> e, ee
	// Generate ephemeral key pair
	state.localEphemeral, err = GenerateNoiseKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ephemeral key: %w", err)
	}

	state.mixHash(state.localEphemeral.PublicKey)

	// Perform DH: ee
	sharedSecret, err = curve25519DH(state.localEphemeral.PrivateKey, state.remoteEphemeral)
	if err != nil {
		return nil, fmt.Errorf("DH ee failed: %w", err)
	}
	state.mixKey(sharedSecret)

	// Generate authentication tag
	authTag := state.generateTag()

	// Send message 2: e + auth tag
	msg2 := make([]byte, 48)
	copy(msg2[:32], state.localEphemeral.PublicKey)
	copy(msg2[32:], authTag)
	if _, err := conn.Write(msg2); err != nil {
		return nil, fmt.Errorf("failed to send message 2: %w", err)
	}

	// Split to transport keys
	session, err := state.split(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to split to transport: %w", err)
	}

	// Clear deadline
	if err := conn.SetDeadline(time.Time{}); err != nil {
		return nil, fmt.Errorf("failed to clear deadline: %w", err)
	}

	return session, nil
}

func (t *NoiseTransport) newHandshakeState(isInitiator bool, remoteStatic []byte) (*NoiseHandshakeState, error) {
	// Initialize with protocol name hash
	protocolHash := sha3.Sum256([]byte(NoiseProtocolName))

	state := &NoiseHandshakeState{
		isInitiator:   isInitiator,
		config:        t.config,
		chainingKey:   protocolHash[:],
		handshakeHash: protocolHash[:],
		localStatic:   t.localKeyPair,
	}

	// Mix in responder's static public key (known in advance for NK pattern)
	if isInitiator && remoteStatic != nil {
		state.remoteStatic = remoteStatic
		state.mixHash(remoteStatic)
	} else if !isInitiator {
		// Responder mixes in their own static public key
		state.mixHash(t.localKeyPair.PublicKey)
	}

	return state, nil
}

func (s *NoiseHandshakeState) mixHash(data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	h := sha3.New256()
	h.Write(s.handshakeHash)
	h.Write(data)
	s.handshakeHash = h.Sum(nil)
}

func (s *NoiseHandshakeState) mixKey(inputKeyMaterial []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Use HKDF to derive new chaining key
	r := hkdf.New(sha3.New256, inputKeyMaterial, s.chainingKey, nil)
	newCK := make([]byte, 32)
	if _, err := io.ReadFull(r, newCK); err != nil {
		panic("HKDF failed: " + err.Error())
	}
	s.chainingKey = newCK
}

func (s *NoiseHandshakeState) generateTag() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()

	h := sha3.New256()
	h.Write(s.handshakeHash)
	h.Write(s.chainingKey)
	return h.Sum(nil)[:16]
}

func (s *NoiseHandshakeState) verifyTag(tag []byte) error {
	expected := s.generateTag()
	if !constantTimeCompare(expected, tag) {
		return errors.New("authentication tag mismatch")
	}
	return nil
}

func (s *NoiseHandshakeState) split(conn net.Conn) (*NoiseSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Derive transport keys using HKDF
	r := hkdf.New(sha3.New256, nil, s.chainingKey, []byte("noise-transport"))
	
	sendKey := make([]byte, 32)
	recvKey := make([]byte, 32)

	if s.isInitiator {
		if _, err := io.ReadFull(r, sendKey); err != nil {
			return nil, err
		}
		if _, err := io.ReadFull(r, recvKey); err != nil {
			return nil, err
		}
	} else {
		if _, err := io.ReadFull(r, recvKey); err != nil {
			return nil, err
		}
		if _, err := io.ReadFull(r, sendKey); err != nil {
			return nil, err
		}
	}

	sendCipher, err := chacha20poly1305.New(sendKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create send cipher: %w", err)
	}

	recvCipher, err := chacha20poly1305.New(recvKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create receive cipher: %w", err)
	}

	// Clear sensitive data
	for i := range sendKey {
		sendKey[i] = 0
	}
	for i := range recvKey {
		recvKey[i] = 0
	}

	return &NoiseSession{
		sendCipher:    sendCipher,
		recvCipher:    recvCipher,
		sendNonce:     0,
		recvNonce:     0,
		handshakeHash: s.handshakeHash,
		conn:          conn,
		readBuf:       make([]byte, 65535+16), // Max message size + auth tag
		writeBuf:      make([]byte, 65535+16),
	}, nil
}

// Read reads an encrypted message from the session.
func (s *NoiseSession) Read(b []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Read length prefix (2 bytes)
	var lengthBuf [2]byte
	if _, err := io.ReadFull(s.conn, lengthBuf[:]); err != nil {
		return 0, err
	}
	msgLen := int(binary.BigEndian.Uint16(lengthBuf[:]))

	if msgLen > len(s.readBuf) {
		return 0, errors.New("message too large")
	}

	// Read encrypted message
	if _, err := io.ReadFull(s.conn, s.readBuf[:msgLen]); err != nil {
		return 0, err
	}

	// Create nonce
	nonce := make([]byte, 12)
	binary.BigEndian.PutUint64(nonce[4:], s.recvNonce)
	s.recvNonce++

	// Decrypt
	plaintext, err := s.recvCipher.Open(nil, nonce, s.readBuf[:msgLen], nil)
	if err != nil {
		return 0, fmt.Errorf("decryption failed: %w", err)
	}

	n := copy(b, plaintext)
	return n, nil
}

// Write writes an encrypted message to the session.
func (s *NoiseSession) Write(b []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(b) > 65535 {
		return 0, errors.New("message too large")
	}

	// Create nonce
	nonce := make([]byte, 12)
	binary.BigEndian.PutUint64(nonce[4:], s.sendNonce)
	s.sendNonce++

	// Encrypt
	ciphertext := s.sendCipher.Seal(nil, nonce, b, nil)

	// Write length prefix
	var lengthBuf [2]byte
	binary.BigEndian.PutUint16(lengthBuf[:], uint16(len(ciphertext)))
	if _, err := s.conn.Write(lengthBuf[:]); err != nil {
		return 0, err
	}

	// Write ciphertext
	if _, err := s.conn.Write(ciphertext); err != nil {
		return 0, err
	}

	return len(b), nil
}

// Close closes the session and underlying connection.
func (s *NoiseSession) Close() error {
	return s.conn.Close()
}

// RemoteAddr returns the remote network address.
func (s *NoiseSession) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

// LocalAddr returns the local network address.
func (s *NoiseSession) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

// SetDeadline sets the read and write deadlines.
func (s *NoiseSession) SetDeadline(t time.Time) error {
	return s.conn.SetDeadline(t)
}

// SetReadDeadline sets the read deadline.
func (s *NoiseSession) SetReadDeadline(t time.Time) error {
	return s.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the write deadline.
func (s *NoiseSession) SetWriteDeadline(t time.Time) error {
	return s.conn.SetWriteDeadline(t)
}

// GetRemotePublicKey returns the remote peer's static public key.
func (s *NoiseSession) GetRemotePublicKey() []byte {
	return s.remotePublicKey
}

// GetHandshakeHash returns the handshake hash for channel binding.
func (s *NoiseSession) GetHandshakeHash() []byte {
	return s.handshakeHash
}

// curve25519DH performs a Curve25519 Diffie-Hellman key exchange.
func curve25519DH(privateKey, publicKey []byte) ([]byte, error) {
	if len(privateKey) != 32 || len(publicKey) != 32 {
		return nil, errors.New("invalid key length")
	}

	var priv, pub [32]byte
	copy(priv[:], privateKey)
	copy(pub[:], publicKey)

	var shared [32]byte
	curve25519.ScalarMult(&shared, &priv, &pub)

	// Clear private key from memory
	for i := range priv {
		priv[i] = 0
	}

	return shared[:], nil
}

// constantTimeCompare performs constant-time comparison of two byte slices.
func constantTimeCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := range a {
		result |= a[i] ^ b[i]
	}
	return result == 0
}
