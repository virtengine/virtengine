package fundauth

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"slices"
	"time"
)

const (
	AuthorizationDomain  = "virtengine.fund-authorization"
	AuthorizationVersion = uint32(1)
)

var (
	ErrInvalidAuthorization = errors.New("invalid fund authorization")
	ErrInvalidSignature     = errors.New("invalid possession signature")
	ErrOutsideBlockBounds   = errors.New("current block outside authorization bounds")
	ErrExpired              = errors.New("authorization expired")
	ErrRegistryMismatch     = errors.New("authorization source does not match registry")
)

type Digest [32]byte

type Amount struct {
	Denom      string
	MinorUnits string
}

type PartyRole uint8

const (
	PartyRoleSender PartyRole = iota + 1
	PartyRoleRecipient
	PartyRolePayer
	PartyRolePayee
	PartyRoleBeneficiary
	PartyRoleOwner
	PartyRoleTreasury
)

type PartyBinding struct {
	Role      PartyRole
	AccountID string
}

type ExpiryKind uint8

const (
	ExpiryAtBlock ExpiryKind = iota + 1
	ExpiryAtUnixTime
)

type ExpiryCoordinate struct {
	Kind  ExpiryKind
	Value uint64
}

type MFAMode uint8

const (
	MFAPossessionOnlyPolicyApproved MFAMode = iota + 1
	MFAEvidenceRequired
)

type EligibilityMode uint8

const (
	EligibilityNotRequired EligibilityMode = iota + 1
	EligibilityEvidenceRequired
)

type FundAuthorization struct {
	Domain               string
	Version              uint32
	ChainID              string
	AccountID            string
	SignerKeyID          string
	SignerKeyEpoch       uint64
	SourceID             string
	TypeURL              string
	Phase                Phase
	Effect               Effect
	MessageDigestHex     string
	Amounts              []Amount
	Parties              []PartyBinding
	CaseDigestHex        string
	OrderDigestHex       string
	ReferenceDigestHex   string
	MFAMode              MFAMode
	MFADigestHex         string
	EligibilityMode      EligibilityMode
	EligibilityDigestHex string
	PolicyDigestHex      string
	NonceDigestHex       string
	IssuedAtBlock        uint64
	IssuedAtUnix         uint64
	LowerBlock           uint64
	UpperBlock           uint64
	Expiry               ExpiryCoordinate
}

type SignedAuthorization struct {
	Authorization FundAuthorization
	Signature     []byte
}

type ResolvedPossessionKey struct {
	AccountID string
	KeyID     string
	Epoch     uint64
	PublicKey ed25519.PublicKey
	Active    bool
}

type Ed25519KeyResolver interface {
	ResolveEd25519(ctx context.Context, accountID, keyID string, epoch uint64) (ResolvedPossessionKey, error)
}

type TransactionBinding struct {
	Domain               string
	Version              uint32
	ChainID              string
	AccountID            string
	SignerKeyID          string
	SignerKeyEpoch       uint64
	SourceID             string
	TypeURL              string
	Phase                Phase
	Effect               Effect
	MessageDigestHex     string
	Amounts              []Amount
	Parties              []PartyBinding
	CaseDigestHex        string
	OrderDigestHex       string
	ReferenceDigestHex   string
	MFAMode              MFAMode
	MFADigestHex         string
	EligibilityMode      EligibilityMode
	EligibilityDigestHex string
	PolicyDigestHex      string
	NonceDigestHex       string
	IssuedAtBlock        uint64
	IssuedAtUnix         uint64
	LowerBlock           uint64
	UpperBlock           uint64
	Expiry               ExpiryCoordinate
	CurrentBlock         uint64
	CurrentTime          time.Time
	MaxClockSkew         time.Duration
	MaxLifetime          time.Duration
}

func Verify(ctx context.Context, signed SignedAuthorization, registry *Registry, resolver Ed25519KeyResolver, binding TransactionBinding) (Digest, error) {
	if err := ctx.Err(); err != nil {
		return Digest{}, err
	}
	if registry == nil || resolver == nil {
		return Digest{}, fmt.Errorf("%w: nil registry or key resolver", ErrInvalidAuthorization)
	}
	if err := validateTransactionBinding(binding); err != nil {
		return Digest{}, err
	}
	auth := signed.Authorization
	descriptor, err := registry.Lookup(auth.SourceID, auth.TypeURL)
	if err != nil || descriptor.Phase != auth.Phase || descriptor.Effect != auth.Effect {
		return Digest{}, ErrRegistryMismatch
	}
	if err := compareBinding(auth, binding); err != nil {
		return Digest{}, err
	}
	if err := descriptor.validateAuthorization(auth); err != nil {
		return Digest{}, err
	}
	if binding.CurrentBlock < auth.LowerBlock || binding.CurrentBlock > auth.UpperBlock {
		return Digest{}, ErrOutsideBlockBounds
	}
	now := uint64(binding.CurrentTime.Unix())
	skew := uint64(binding.MaxClockSkew / time.Second)
	maxLifetime := uint64(binding.MaxLifetime / time.Second)
	if auth.IssuedAtUnix > now && auth.IssuedAtUnix-now > skew {
		return Digest{}, fmt.Errorf("%w: authorization issued in future", ErrInvalidAuthorization)
	}
	if now > auth.IssuedAtUnix && now-auth.IssuedAtUnix > maxLifetime {
		return Digest{}, ErrExpired
	}
	switch auth.Expiry.Kind {
	case ExpiryAtBlock:
		if binding.CurrentBlock > auth.Expiry.Value {
			return Digest{}, ErrExpired
		}
	case ExpiryAtUnixTime:
		if auth.Expiry.Value < auth.IssuedAtUnix || auth.Expiry.Value-auth.IssuedAtUnix > maxLifetime || now > auth.Expiry.Value {
			return Digest{}, ErrExpired
		}
	default:
		return Digest{}, fmt.Errorf("%w: expiry kind", ErrInvalidAuthorization)
	}
	signBytes, digest, err := CanonicalSignBytes(auth)
	if err != nil {
		return Digest{}, err
	}
	resolved, err := resolver.ResolveEd25519(ctx, auth.AccountID, auth.SignerKeyID, auth.SignerKeyEpoch)
	if err != nil {
		return Digest{}, fmt.Errorf("resolve possession key: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return Digest{}, err
	}
	if !resolved.Active || resolved.AccountID != auth.AccountID || resolved.KeyID != auth.SignerKeyID || resolved.Epoch != auth.SignerKeyEpoch {
		return Digest{}, fmt.Errorf("%w: resolved possession key identity", ErrInvalidAuthorization)
	}
	if len(resolved.PublicKey) != ed25519.PublicKeySize || len(signed.Signature) != ed25519.SignatureSize || !ed25519.Verify(resolved.PublicKey, signBytes, signed.Signature) {
		return Digest{}, ErrInvalidSignature
	}
	return digest, nil
}

func validateTransactionBinding(binding TransactionBinding) error {
	if binding.Domain != AuthorizationDomain || binding.Version != AuthorizationVersion || binding.ChainID == "" || binding.AccountID == "" || binding.SignerKeyID == "" || binding.SignerKeyEpoch == 0 || binding.SourceID == "" || binding.MessageDigestHex == "" || binding.PolicyDigestHex == "" || binding.NonceDigestHex == "" || binding.IssuedAtBlock == 0 || binding.IssuedAtUnix == 0 || binding.LowerBlock == 0 || binding.UpperBlock == 0 || binding.Expiry.Value == 0 || binding.CurrentBlock == 0 || binding.CurrentTime.IsZero() || binding.CurrentTime.Unix() < 0 || binding.MaxClockSkew <= 0 || binding.MaxLifetime <= 0 {
		return fmt.Errorf("%w: incomplete transaction binding", ErrInvalidAuthorization)
	}
	return nil
}

func compareBinding(auth FundAuthorization, binding TransactionBinding) error {
	if auth.Domain != binding.Domain || auth.Version != binding.Version || auth.ChainID != binding.ChainID || auth.AccountID != binding.AccountID || auth.SignerKeyID != binding.SignerKeyID || auth.SignerKeyEpoch != binding.SignerKeyEpoch || auth.SourceID != binding.SourceID || auth.TypeURL != binding.TypeURL || auth.Phase != binding.Phase || auth.Effect != binding.Effect || auth.MessageDigestHex != binding.MessageDigestHex || !slices.Equal(auth.Amounts, binding.Amounts) || !slices.Equal(auth.Parties, binding.Parties) || auth.CaseDigestHex != binding.CaseDigestHex || auth.OrderDigestHex != binding.OrderDigestHex || auth.ReferenceDigestHex != binding.ReferenceDigestHex || auth.MFAMode != binding.MFAMode || auth.MFADigestHex != binding.MFADigestHex || auth.EligibilityMode != binding.EligibilityMode || auth.EligibilityDigestHex != binding.EligibilityDigestHex || auth.PolicyDigestHex != binding.PolicyDigestHex || auth.NonceDigestHex != binding.NonceDigestHex || auth.IssuedAtBlock != binding.IssuedAtBlock || auth.IssuedAtUnix != binding.IssuedAtUnix || auth.LowerBlock != binding.LowerBlock || auth.UpperBlock != binding.UpperBlock || auth.Expiry != binding.Expiry {
		return fmt.Errorf("%w: transaction binding mismatch", ErrInvalidAuthorization)
	}
	return nil
}

// AtomicAuthorizationConsumer is the persistence contract for T5-18 keeper
// implementation. This package intentionally provides no production consumer.
type AtomicAuthorizationConsumer interface {
	KeeperRequired() bool
	WithAuthorization(ctx context.Context, accountID string, nonceDigest, authDigest Digest, protected func(context.Context) error) error
}

func VerifyAndConsume(ctx context.Context, signed SignedAuthorization, registry *Registry, resolver Ed25519KeyResolver, binding TransactionBinding, consumer AtomicAuthorizationConsumer, protected func(context.Context) error) error {
	if consumer == nil || !consumer.KeeperRequired() || protected == nil {
		return fmt.Errorf("%w: nil consumer or callback", ErrInvalidAuthorization)
	}
	authDigest, err := Verify(ctx, signed, registry, resolver, binding)
	if err != nil {
		return err
	}
	nonceDigest, err := parseRequiredDigest(signed.Authorization.NonceDigestHex, "nonce")
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return consumer.WithAuthorization(ctx, signed.Authorization.AccountID, nonceDigest, authDigest, protected)
}
