package testutil

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	cosmosclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdktest "github.com/cosmos/cosmos-sdk/testutil"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	pcmd "github.com/virtengine/virtengine/provider/cmd"
	testutilcli "github.com/virtengine/virtengine/testutil/cli"
	mtypes "github.com/virtengine/virtengine/x/market/types"
)

const (
	TestClusterPublicHostname   = "e2e.test"
	TestClusterNodePortQuantity = 100
)

/*
TestSendManifest for integration testing
this is similar to cli command exampled below
virtengine provider send-manifest --owner <address> \
	--dseq 7 \
	--provider <address> ./../_run/kube/deployment.yaml \
	--home=/tmp/virtengine_integration_TestE2EApp_324892307/.virtenginectl --node=tcp://0.0.0.0:41863
*/
func TestSendManifest(clientCtx client.Context, id mtypes.BidID, sdlPath string, extraArgs ...string) (sdktest.BufferWriter, error) {
	args := []string{
		fmt.Sprintf("--dseq=%v", id.DSeq),
		fmt.Sprintf("--provider=%s", id.Provider),
	}
	args = append(args, sdlPath)
	args = append(args, extraArgs...)
	fmt.Printf("%v\n", args)
	return testutilcli.ExecTestCLICmd(clientCtx, pcmd.SendManifestCmd(), args...)
}

func TestLeaseShell(clientCtx client.Context, extraArgs []string, lID mtypes.LeaseID, replicaIndex int, tty bool, stdin bool, serviceName string, cmd ...string) (sdktest.BufferWriter, error) {
	args := []string{
		fmt.Sprintf("--provider=%s", lID.Provider),
		fmt.Sprintf("--replica-index=%d", replicaIndex),
		fmt.Sprintf("--dseq=%v", lID.DSeq),
		fmt.Sprintf("--gseq=%v", lID.GSeq),
	}
	if tty {
		args = append(args, "--tty")
	}
	if stdin {
		args = append(args, "--stdin")
	}
	args = append(args, extraArgs...)
	args = append(args, serviceName)
	args = append(args, cmd...)
	fmt.Printf("%v\n", args)
	return testutilcli.ExecTestCLICmd(clientCtx, pcmd.LeaseShellCmd(), args...)
}

// RunLocalProvider wraps up the Provider cobra command for testing and supplies
// new default values to the flags.
// prev: virtenginectl provider run --from=foo --cluster-k8s --gateway-listen-address=localhost:39729 --home=/tmp/virtengine_integration_TestE2EApp_324892307/.virtenginectl --node=tcp://0.0.0.0:41863 --keyring-backend test
func RunLocalProvider(clientCtx cosmosclient.Context, chainID, nodeRPC, virtengineHome, from, gatewayListenAddress string, extraArgs ...string) (sdktest.BufferWriter, error) {
	cmd := pcmd.RunCmd()
	// Flags added because command not being wrapped by the Tendermint's PrepareMainCmd()
	cmd.PersistentFlags().StringP(tmcli.HomeFlag, "", virtengineHome, "directory for config and data")
	cmd.PersistentFlags().Bool(tmcli.TraceFlag, false, "print out full stack trace on errors")

	args := []string{
		fmt.Sprintf("--%s", pcmd.FlagClusterK8s),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, chainID),
		fmt.Sprintf("--%s=%s", flags.FlagNode, nodeRPC),
		fmt.Sprintf("--%s=%s", flags.FlagHome, virtengineHome),
		fmt.Sprintf("--from=%s", from),
		fmt.Sprintf("--%s=%s", pcmd.FlagGatewayListenAddress, gatewayListenAddress),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
		fmt.Sprintf("--%s=%s", pcmd.FlagClusterPublicHostname, TestClusterPublicHostname),
		fmt.Sprintf("--%s=%d", pcmd.FlagClusterNodePortQuantity, TestClusterNodePortQuantity),
		fmt.Sprintf("--%s=%s", pcmd.FlagBidPricingStrategy, "randomRange"),
	}

	args = append(args, extraArgs...)

	return testutilcli.ExecTestCLICmd(clientCtx, cmd, args...)
}
