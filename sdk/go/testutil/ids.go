package testutil

import (
	cryptorand "crypto/rand"
	"crypto/sha256"
	"math/rand" //nolint:gosec // G404: test helpers use weak random for non-security data
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	dtypes "github.com/virtengine/virtengine/sdk/go/node/deployment/v1"
	mtypes "github.com/virtengine/virtengine/sdk/go/node/market/v1"
)

func DeploymentID(t testing.TB) dtypes.DeploymentID {
	t.Helper()
	return dtypes.DeploymentID{
		Owner: AccAddress(t).String(),
		DSeq:  uint64(rand.Uint32()), //nolint:gosec // G404: test helper randomness is non-security-sensitive
	}
}

func DeploymentIDForAccount(t testing.TB, addr sdk.Address) dtypes.DeploymentID {
	t.Helper()
	return dtypes.DeploymentID{
		Owner: addr.String(),
		DSeq:  uint64(rand.Uint32()), //nolint:gosec // G404: test helper randomness is non-security-sensitive
	}
}

// DeploymentVersion provides a random sha256 sum for simulating Deployments.
func DeploymentVersion(t testing.TB) []byte {
	t.Helper()
	src := make([]byte, 128)
	_, err := cryptorand.Read(src)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(src)
	return sum[:]
}

func GroupID(t testing.TB) dtypes.GroupID {
	t.Helper()
	return dtypes.MakeGroupID(DeploymentID(t), rand.Uint32()) //nolint:gosec // G404: test helper randomness is non-security-sensitive
}

func GroupIDForAccount(t testing.TB, addr sdk.Address) dtypes.GroupID {
	t.Helper()
	return dtypes.MakeGroupID(DeploymentIDForAccount(t, addr), rand.Uint32()) //nolint:gosec // G404: test helper randomness is non-security-sensitive
}

func OrderID(t testing.TB) mtypes.OrderID {
	t.Helper()
	return mtypes.MakeOrderID(GroupID(t), rand.Uint32()) //nolint:gosec // G404: test helper randomness is non-security-sensitive
}

func OrderIDForAccount(t testing.TB, addr sdk.Address) mtypes.OrderID {
	t.Helper()
	return mtypes.MakeOrderID(GroupIDForAccount(t, addr), rand.Uint32()) //nolint:gosec // G404: test helper randomness is non-security-sensitive
}

func BidID(t testing.TB) mtypes.BidID {
	t.Helper()
	return mtypes.MakeBidID(OrderID(t), AccAddress(t))
}

func BidIDForAccount(t testing.TB, owner, provider sdk.Address) mtypes.BidID {
	t.Helper()
	return mtypes.MakeBidID(OrderIDForAccount(t, owner), provider.Bytes())
}

func LeaseID(t testing.TB) mtypes.LeaseID {
	t.Helper()
	return mtypes.MakeLeaseID(BidID(t))
}

func LeaseIDForAccount(t testing.TB, owner, provider sdk.Address) mtypes.LeaseID {
	t.Helper()
	return mtypes.MakeLeaseID(BidIDForAccount(t, owner, provider))
}
