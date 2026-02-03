package keeper

import (
	"context"
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	encryptionv1 "github.com/virtengine/virtengine/sdk/go/node/encryption/v1"
	"github.com/virtengine/virtengine/x/encryption/types"
)

// Error message constant
const errMsgEmptyRequest = "empty request"

// GRPCQuerier implements the gRPC query interface with proper context handling
type GRPCQuerier struct {
	Keeper
}

var _ types.QueryServer = GRPCQuerier{}

// RecipientKey returns all keys registered for an address
func (q GRPCQuerier) RecipientKey(c context.Context, req *types.QueryRecipientKeyRequest) (*types.QueryRecipientKeyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	ctx := sdk.UnwrapSDKContext(c)

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(err.Error())
	}

	localKeys := q.GetRecipientKeys(ctx, addr)

	// Convert local keys to proto type
	keys := make([]encryptionv1.RecipientKeyRecord, len(localKeys))
	for i, k := range localKeys {
		keys[i] = encryptionv1.RecipientKeyRecord{
			Address:        k.Address,
			PublicKey:      k.PublicKey,
			KeyFingerprint: k.KeyFingerprint,
			AlgorithmId:    k.AlgorithmID,
			RegisteredAt:   k.RegisteredAt,
			RevokedAt:      k.RevokedAt,
			Label:          k.Label,
		}
	}

	return &types.QueryRecipientKeyResponse{
		Keys: keys,
	}, nil
}

// KeyByFingerprint returns a key by its fingerprint
func (q GRPCQuerier) KeyByFingerprint(c context.Context, req *types.QueryKeyByFingerprintRequest) (*types.QueryKeyByFingerprintResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	ctx := sdk.UnwrapSDKContext(c)

	key, found := q.GetRecipientKeyByFingerprint(ctx, req.Fingerprint)
	if !found {
		return nil, types.ErrKeyNotFound.Wrapf("key with fingerprint %s not found", req.Fingerprint)
	}

	return &types.QueryKeyByFingerprintResponse{
		Key: &encryptionv1.RecipientKeyRecord{
			Address:        key.Address,
			PublicKey:      key.PublicKey,
			KeyFingerprint: key.KeyFingerprint,
			AlgorithmId:    key.AlgorithmID,
			RegisteredAt:   key.RegisteredAt,
			RevokedAt:      key.RevokedAt,
			Label:          key.Label,
		},
	}, nil
}

// Params returns the module parameters
func (q GRPCQuerier) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	ctx := sdk.UnwrapSDKContext(c)
	params := q.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}

// Algorithms returns the supported algorithms
func (q GRPCQuerier) Algorithms(c context.Context, req *types.QueryAlgorithmsRequest) (*types.QueryAlgorithmsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	var algorithms []encryptionv1.AlgorithmInfo
	for _, algID := range types.SupportedAlgorithms() {
		info, err := types.GetAlgorithmInfo(algID)
		if err == nil {
			algorithms = append(algorithms, encryptionv1.AlgorithmInfo{
				Id:          info.ID,
				Version:     info.Version,
				Description: info.Description,
				KeySize:     safeInt32FromInt(info.KeySize),
				NonceSize:   safeInt32FromInt(info.NonceSize),
				//nolint:staticcheck // Deprecated field is required for compatibility with existing clients.
				Deprecated: info.Deprecated,
			})
		}
	}

	return &types.QueryAlgorithmsResponse{
		Algorithms: algorithms,
	}, nil
}

// ValidateEnvelope validates an envelope structure
func (q GRPCQuerier) ValidateEnvelope(c context.Context, req *types.QueryValidateEnvelopeRequest) (*types.QueryValidateEnvelopeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	ctx := sdk.UnwrapSDKContext(c)

	envelope := &req.Envelope
	if len(envelope.RecipientKeyIds) > math.MaxInt32 {
		return nil, status.Error(codes.InvalidArgument, "recipient count exceeds int32")
	}
	response := &types.QueryValidateEnvelopeResponse{
		Valid:          true,
		RecipientCount: safeInt32FromInt(len(envelope.RecipientKeyIds)),
		Algorithm:      envelope.AlgorithmId,
	}

	// Convert proto envelope to local type for validation
	localEnvelope := convertProtoEnvelopeToLocal(envelope)

	// Validate envelope structure
	if err := q.Keeper.ValidateEnvelope(ctx, localEnvelope); err != nil {
		response.Valid = false
		response.Error = err.Error()
		return response, nil
	}

	// Validate recipients
	missingKeys, err := q.ValidateEnvelopeRecipients(ctx, localEnvelope)
	if err != nil {
		response.Valid = false
		response.Error = err.Error()
		return response, nil
	}

	response.MissingKeys = missingKeys
	response.AllKeysRegistered = len(missingKeys) == 0

	// Signature validation
	response.SignatureValid = len(envelope.SenderSignature) > 0

	return response, nil
}

func safeInt32FromInt(value int) int32 {
	if value < math.MinInt32 {
		return math.MinInt32
	}
	if value > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(value)
}

// convertProtoEnvelopeToLocal converts a proto envelope to the local type
func convertProtoEnvelopeToLocal(pb *encryptionv1.EncryptedPayloadEnvelope) *types.EncryptedPayloadEnvelope {
	wrappedKeys := make([]types.WrappedKeyEntry, len(pb.WrappedKeys))
	for i, wk := range pb.WrappedKeys {
		wrappedKeys[i] = types.WrappedKeyEntry{
			RecipientID:     wk.RecipientId,
			WrappedKey:      wk.WrappedKey,
			Algorithm:       wk.Algorithm,
			EphemeralPubKey: wk.EphemeralPubKey,
		}
	}

	return &types.EncryptedPayloadEnvelope{
		Version:             pb.Version,
		AlgorithmID:         pb.AlgorithmId,
		AlgorithmVersion:    pb.AlgorithmVersion,
		RecipientKeyIDs:     pb.RecipientKeyIds,
		RecipientPublicKeys: pb.RecipientPublicKeys,
		EncryptedKeys:       pb.EncryptedKeys,
		WrappedKeys:         wrappedKeys,
		Nonce:               pb.Nonce,
		Ciphertext:          pb.Ciphertext,
		SenderSignature:     pb.SenderSignature,
		SenderPubKey:        pb.SenderPubKey,
		Metadata:            pb.Metadata,
	}
}
