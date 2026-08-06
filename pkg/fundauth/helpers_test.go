package fundauth

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"
)

func testDigest(label string) string {
	digest := sha256.Sum256([]byte(label))
	return hex.EncodeToString(digest[:])
}

func validAuthorization() FundAuthorization {
	return FundAuthorization{
		Domain:               AuthorizationDomain,
		Version:              AuthorizationVersion,
		ChainID:              "virtengine-test-1",
		AccountID:            "account:alice",
		SignerKeyID:          "device-key:primary",
		SignerKeyEpoch:       7,
		SourceID:             "/cosmos.bank.v1beta1.MsgSend",
		TypeURL:              "/cosmos.bank.v1beta1.MsgSend",
		Phase:                PhaseImmediate,
		Effect:               EffectTransfer,
		MessageDigestHex:     testDigest("message"),
		Amounts:              []Amount{{Denom: "act", MinorUnits: "42"}, {Denom: "uve", MinorUnits: "1000000"}},
		Parties:              []PartyBinding{{Role: PartyRoleSender, AccountID: "account:alice"}, {Role: PartyRoleRecipient, AccountID: "account:bob"}},
		CaseDigestHex:        testDigest("case"),
		OrderDigestHex:       testDigest("order"),
		ReferenceDigestHex:   testDigest("reference"),
		MFAMode:              MFAEvidenceRequired,
		MFADigestHex:         testDigest("mfa"),
		EligibilityMode:      EligibilityEvidenceRequired,
		EligibilityDigestHex: testDigest("eligibility"),
		PolicyDigestHex:      testDigest("policy"),
		NonceDigestHex:       testDigest("nonce"),
		IssuedAtBlock:        100,
		IssuedAtUnix:         1_799_999_900,
		LowerBlock:           100,
		UpperBlock:           200,
		Expiry:               ExpiryCoordinate{Kind: ExpiryAtBlock, Value: 180},
	}
}

func cloneAuthorization(auth FundAuthorization) FundAuthorization {
	auth.Amounts = append([]Amount(nil), auth.Amounts...)
	auth.Parties = append([]PartyBinding(nil), auth.Parties...)
	return auth
}

type fixedResolver struct {
	accountID string
	keyID     string
	epoch     uint64
	publicKey ed25519.PublicKey
	active    bool
}

func (resolver fixedResolver) ResolveEd25519(_ context.Context, accountID, keyID string, epoch uint64) (ResolvedPossessionKey, error) {
	if accountID != resolver.accountID || keyID != resolver.keyID || epoch != resolver.epoch {
		return ResolvedPossessionKey{}, errors.New("unbound key coordinates")
	}
	return ResolvedPossessionKey{AccountID: resolver.accountID, KeyID: resolver.keyID, Epoch: resolver.epoch, PublicKey: append(ed25519.PublicKey(nil), resolver.publicKey...), Active: resolver.active}, nil
}

func signedFixture(t *testing.T) (SignedAuthorization, fixedResolver, ed25519.PrivateKey) {
	t.Helper()
	seed := sha256.Sum256([]byte("fundauth deterministic test key"))
	privateKey := ed25519.NewKeyFromSeed(seed[:])
	auth := validAuthorization()
	signBytes, _, err := CanonicalSignBytes(auth)
	if err != nil {
		t.Fatal(err)
	}
	return SignedAuthorization{Authorization: auth, Signature: ed25519.Sign(privateKey, signBytes)}, fixedResolver{
		accountID: auth.AccountID,
		keyID:     auth.SignerKeyID,
		epoch:     auth.SignerKeyEpoch,
		publicKey: privateKey.Public().(ed25519.PublicKey),
		active:    true,
	}, privateKey
}

func verifyOptions(auth FundAuthorization) TransactionBinding {
	return TransactionBinding{
		Domain: auth.Domain, Version: auth.Version, ChainID: auth.ChainID, AccountID: auth.AccountID,
		SignerKeyID: auth.SignerKeyID, SignerKeyEpoch: auth.SignerKeyEpoch, SourceID: auth.SourceID, TypeURL: auth.TypeURL,
		Phase: auth.Phase, Effect: auth.Effect, MessageDigestHex: auth.MessageDigestHex,
		Amounts: append([]Amount(nil), auth.Amounts...), Parties: append([]PartyBinding(nil), auth.Parties...),
		CaseDigestHex: auth.CaseDigestHex, OrderDigestHex: auth.OrderDigestHex, ReferenceDigestHex: auth.ReferenceDigestHex,
		MFAMode: auth.MFAMode, MFADigestHex: auth.MFADigestHex, EligibilityMode: auth.EligibilityMode,
		EligibilityDigestHex: auth.EligibilityDigestHex, PolicyDigestHex: auth.PolicyDigestHex, NonceDigestHex: auth.NonceDigestHex,
		IssuedAtBlock: auth.IssuedAtBlock, IssuedAtUnix: auth.IssuedAtUnix, LowerBlock: auth.LowerBlock, UpperBlock: auth.UpperBlock, Expiry: auth.Expiry,
		CurrentBlock: 150, CurrentTime: time.Unix(1_800_000_000, 0), MaxClockSkew: 30 * time.Second, MaxLifetime: time.Hour,
	}
}
