//go:build e2e.integration

// Package consensus contains in-process multi-validator consensus integration tests.
package consensus

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	coreheader "cosmossdk.io/core/header"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/app"
	apptypes "github.com/virtengine/virtengine/app/types"
	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
	"github.com/virtengine/virtengine/testutil"
	networktest "github.com/virtengine/virtengine/testutil/network"
	utypes "github.com/virtengine/virtengine/upgrades/types"
)

const (
	task84AUpgradeHeight      = int64(1)
	task84APostActivation     = int64(500)
	task84ADefaultLoadTxCount = 5000
	task84ATxGas              = uint64(100_000)
	task84ACommitTimeout      = 5 * time.Millisecond
	task84AConsensusRoundTime = 2 * time.Second
	task84ATestTimeout        = 15 * time.Minute
)

var task84ATimeZones = []string{"UTC", "Pacific/Auckland", "America/Los_Angeles", "Asia/Kolkata"}

type processObservation struct {
	Validator int
	Height    int64
	TxCount   int
	Duration  time.Duration
	Status    abci.ResponseProcessProposal_ProposalStatus
}

type processCollector struct {
	mu             sync.Mutex
	observations   []processObservation
	malicious      []processObservation
	maxPrepared    int
	injected       atomic.Bool
	injectedHeight atomic.Int64
}

func (c *processCollector) record(observation processObservation, malicious bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.observations = append(c.observations, observation)
	if malicious {
		c.malicious = append(c.malicious, observation)
	}
}

func (c *processCollector) snapshot() ([]processObservation, []processObservation) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]processObservation(nil), c.observations...), append([]processObservation(nil), c.malicious...)
}

func (c *processCollector) prepared(txCount int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if txCount > c.maxPrepared {
		c.maxPrepared = txCount
	}
}

func (c *processCollector) maxPreparedTxs() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.maxPrepared
}

type observedApplication struct {
	servertypes.Application
	validator int
	collector *processCollector
	loadTxs   *atomic.Pointer[[][]byte]
}

func (a *observedApplication) InitChain(req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	response, err := a.Application.InitChain(req)
	if err != nil {
		return nil, err
	}
	validatorApp, ok := validatorAppFromApplication(a.Application)
	if !ok {
		return nil, fmt.Errorf("unexpected app type %T", a.Application)
	}
	ctx := validatorApp.NewContextLegacy(false, cmtproto.Header{
		ChainID: req.ChainId,
		Height:  task84AUpgradeHeight,
		Time:    req.Time,
	}).WithHeaderInfo(coreheader.Info{
		ChainID: req.ChainId,
		Height:  task84AUpgradeHeight,
		Time:    req.Time,
	})
	if err := validatorApp.Keepers.Cosmos.Upgrade.ApplyUpgrade(ctx, upgradetypes.Plan{
		Name:   utypes.ConsensusAdmissionUpgradeName,
		Height: task84AUpgradeHeight,
	}); err != nil {
		return nil, fmt.Errorf("execute %s during InitChain: %w", utypes.ConsensusAdmissionUpgradeName, err)
	}
	params, err := validatorApp.Keepers.Cosmos.ConsensusParams.ParamsStore.Get(ctx)
	if err != nil {
		return nil, err
	}
	response.ConsensusParams = &params
	return response, nil
}

func (a *observedApplication) PrepareProposal(req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
	if req != nil && a.loadTxs != nil {
		if load := a.loadTxs.Load(); load != nil && len(req.Txs) == 0 {
			reqCopy := *req
			reqCopy.Txs = cloneTxBytes(*load)
			req = &reqCopy
		}
	}
	response, err := a.Application.PrepareProposal(req)
	if response != nil && (req == nil || req.Height != a.collector.injectedHeight.Load()) {
		a.collector.prepared(len(response.Txs))
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "TASK84A prepare validator=%d height=%d error=%v\n", a.validator, requestHeight(req), err)
	}
	if os.Getenv("VE_TASK84A_DEBUG") == "1" && req != nil && req.Height >= task84AUpgradeHeight+1 {
		responseTxs := -1
		if response != nil {
			responseTxs = len(response.Txs)
		}
		fmt.Fprintf(os.Stderr, "TASK84A prepare validator=%d height=%d candidate_txs=%d response_txs=%d first_system=%t\n", a.validator, req.Height, len(req.Txs), responseTxs, response != nil && len(response.Txs) > 0 && isConsensusSystemTx(a.Application, response.Txs[0]))
	}
	if err != nil || response == nil || req == nil || len(req.Txs) == 0 || a.collector.injected.Load() {
		return response, err
	}
	if !a.collector.injected.CompareAndSwap(false, true) {
		return response, err
	}
	a.collector.injectedHeight.Store(req.Height)
	// Omit the required index-zero carrier from the genuine proposal response.
	// Honest validators receive these bytes and reject them in ProcessProposal.
	if len(response.Txs) > 1 {
		response.Txs = append([][]byte(nil), response.Txs[1:]...)
	} else {
		response.Txs = nil
	}
	return response, err
}

func cloneTxBytes(txs [][]byte) [][]byte {
	cloned := make([][]byte, len(txs))
	for index := range txs {
		cloned[index] = bytes.Clone(txs[index])
	}
	return cloned
}

func (a *observedApplication) ProcessProposal(req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
	// Wall-clock duration is observational only: it is recorded after the wrapped
	// decision and is never read by the application or proposal policy.
	start := time.Now()
	response, err := a.Application.ProcessProposal(req)
	duration := time.Since(start)
	if req != nil {
		status := abci.ResponseProcessProposal_UNKNOWN
		if response != nil {
			status = response.Status
		}
		injectedHeight := a.collector.injectedHeight.Load()
		malicious := injectedHeight > 0 && req.Height == injectedHeight && (len(req.Txs) == 0 || !isConsensusSystemTx(a.Application, req.Txs[0]))
		a.collector.record(processObservation{
			Validator: a.validator,
			Height:    req.Height,
			TxCount:   len(req.Txs),
			Duration:  duration,
			Status:    status,
		}, malicious)
		if a.loadTxs != nil && injectedHeight > 0 && req.Height > injectedHeight && len(req.Txs) > 1 && status == abci.ResponseProcessProposal_REJECT {
			a.loadTxs.Store(nil)
		}
		if err != nil || status == abci.ResponseProcessProposal_REJECT {
			fmt.Fprintf(os.Stderr, "TASK84A process validator=%d height=%d txs=%d status=%s error=%v\n", a.validator, req.Height, len(req.Txs), status.String(), err)
		}
	}
	return response, err
}

func (a *observedApplication) FinalizeBlock(req *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	response, err := a.Application.FinalizeBlock(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "TASK84A finalize validator=%d height=%d error=%v\n", a.validator, requestFinalizeHeight(req), err)
	} else if a.loadTxs != nil && req != nil && a.collector.injectedHeight.Load() > 0 && req.Height > a.collector.injectedHeight.Load() && len(req.Txs) > 1 {
		a.loadTxs.Store(nil)
	}
	return response, err
}

func requestHeight(req *abci.RequestPrepareProposal) int64 {
	if req == nil {
		return 0
	}
	return req.Height
}

func requestFinalizeHeight(req *abci.RequestFinalizeBlock) int64 {
	if req == nil {
		return 0
	}
	return req.Height
}

// TestTask84AFourValidatorAdversarialLoad is opt-in because it starts four
// independent CometBFT nodes/apps and advances 500 strict-admission blocks.
func TestTask84AFourValidatorAdversarialLoad(t *testing.T) {
	if os.Getenv("VE_RUN_TASK84A_CONSENSUS") != "1" {
		t.Skip("set VE_RUN_TASK84A_CONSENSUS=1 to run the 500-block consensus evidence test")
	}

	testStarted := time.Now()
	collector := &processCollector{}
	loadSource := &atomic.Pointer[[][]byte]{}
	loadTxCount := envInt(t, "VE_TASK84A_LOAD_TXS", task84ADefaultLoadTxCount)
	require.LessOrEqual(t, loadTxCount, app.DefaultProposalLimits().MaxTxCount)
	chainID := "task-84a-consensus-evidence"

	fixtureFactory := func(_ ...networktest.TestnetFixtureOption) networktest.TestFixture {
		return testutil.NewTestNetworkFixture(
			networktest.WithApplicationWrapper(func(val networktest.ValidatorI, rawApp servertypes.Application) servertypes.Application {
				return &observedApplication{
					Application: rawApp,
					validator:   val.GetIndex(),
					collector:   collector,
					loadTxs:     loadSource,
				}
			}),
		)
	}

	cfg := networktest.DefaultConfig(fixtureFactory)
	cfg.ChainID = chainID
	cfg.NumValidators = 4
	cfg.TimeoutCommit = task84ACommitTimeout
	cfg.EnableLogging = os.Getenv("VE_TASK84A_DEBUG") == "1"
	cfg.AdditionalAccounts = (loadTxCount + int(apptypes.DefaultMaxTxPerBlockPerAccount) - 1) / int(apptypes.DefaultMaxTxPerBlockPerAccount)
	cfg.AppOptions = func(index int, home string) servertypes.AppOptions {
		rateLimits := apptypes.DefaultRateLimitParams()
		rateLimits.MaxTotalTxPerBlock = uint64(loadTxCount) //nolint:gosec // load count is positive and bounded to 5,000 above
		return simtestutil.AppOptionsMap{
			"home":                                home,
			"test_rate_limit_params":              rateLimits,
			"test_enable_unordered_transactions":  true,
			"test_checktx_prevalidated_proposals": true,
			"test_proposal_observer": func(stage string, err error) {
				if os.Getenv("VE_TASK84A_DEBUG") == "1" {
					fmt.Fprintf(os.Stderr, "TASK84A proposal-observer validator=%d stage=%s error=%v\n", index, stage, err)
				}
			},
		}
	}
	cfg.ConfigureNode = func(index int, ctx *server.Context, _ *serverconfig.Config) {
		ctx.Config.Consensus.TimeoutPropose = task84AConsensusRoundTime
		ctx.Config.Consensus.TimeoutProposeDelta = 100 * time.Millisecond
		ctx.Config.Consensus.TimeoutPrevote = 100 * time.Millisecond
		ctx.Config.Consensus.TimeoutPrevoteDelta = 50 * time.Millisecond
		ctx.Config.Consensus.TimeoutPrecommit = 100 * time.Millisecond
		ctx.Config.Consensus.TimeoutPrecommitDelta = 50 * time.Millisecond
		ctx.Config.Consensus.SkipTimeoutCommit = true
		ctx.Config.Mempool.Size = loadTxCount
		ctx.Config.Mempool.CacheSize = loadTxCount * 2
		ctx.Config.Mempool.Recheck = false
		ctx.Config.TxIndex.Indexer = "null"
		ctx.Viper.Set("task84a.time_zone", task84ATimeZones[index])
		ctx.Viper.Set("task84a.locale", []string{"C", "en_NZ", "en_US", "tr_TR"}[index])
		ctx.Viper.Set("task84a.host_clock_offset", []string{"-12h", "+13h", "-7h", "+5h30m"}[index])
	}

	net := networktest.New(t, cfg)
	t.Cleanup(net.Cleanup)
	ctx, cancel := context.WithTimeout(context.Background(), task84ATestTimeout)
	defer cancel()

	requireAllValidatorsAtLeast(t, ctx, net, task84AUpgradeHeight+1)
	requireStrictActivation(t, net, task84AUpgradeHeight+1)
	loadTxs := buildSignedLoadTxs(t, net, cfg, loadTxCount)
	// The production proposal path normally repeats ante checks after CheckTx.
	// This bounded stress batch first passes the real CheckTx path, then the
	// test-only app option avoids mutating proposal check-state 5,000 times.
	// Canonical/count/byte/gas/system validation and normal Finalize ante remain.
	validateLoadThroughCheckTx(t, net, loadTxs)
	loadSource.Store(&loadTxs)

	targetHeight := task84AUpgradeHeight + task84APostActivation
	requireAllValidatorsAtLeast(t, ctx, net, targetHeight)
	heights, hashes := validatorEvidenceAtHeight(t, ctx, net, targetHeight)
	for index := 1; index < len(hashes); index++ {
		require.Equal(t, hashes[0], hashes[index], "validator %d app hash differs at target height", index)
	}

	block, err := net.Validators[0].RPCClient.Block(ctx, &targetHeight)
	require.NoError(t, err)
	require.NotEmpty(t, block.Block.Txs)
	require.True(t, isConsensusSystemTx(netApp(t, net, 0), block.Block.Txs[0]), "target block is not under strict carrier admission")

	observations, malicious := collector.snapshot()
	assertMaliciousProposalRejected(t, malicious)
	durations, perValidator := strictProcessDurations(observations, task84AUpgradeHeight+1, targetHeight)
	require.NotEmpty(t, durations)
	p99 := percentile99(durations)
	latencyLimit := minDuration(time.Second, task84AConsensusRoundTime/2)
	require.Less(t, p99, latencyLimit)

	maxBlockTxs, totalOrdinaryTxs := blockLoadEvidence(t, ctx, net.Validators[0], netApp(t, net, 0), task84AUpgradeHeight+1, targetHeight)
	require.Equal(t, loadTxCount+1, collector.maxPreparedTxs(), "an honest proposer must prepare the full configured load")
	require.Equal(t, loadTxCount+1, maxBlockTxs, "the loaded strict block must include one system tx plus all ordinary load txs")
	require.GreaterOrEqual(t, totalOrdinaryTxs, loadTxCount, "at least one full signed transaction batch must commit")
	for validator := 0; validator < cfg.NumValidators; validator++ {
		require.Positive(t, perValidator[validator], "validator %d did not observe strict ProcessProposal", validator)
	}

	t.Logf("TASK84A_RESULT duration=%s validators=4 upgrade_height=%d activation_height=%d target_height=%d post_activation_blocks=%d signed_load_txs=%d max_txs_in_block=%d process_samples=%d p99=%s latency_limit=%s heights=%v app_hash=%X malicious_rejections=%d time_zones=%v",
		time.Since(testStarted).Round(time.Millisecond), task84AUpgradeHeight, task84AUpgradeHeight+1, targetHeight,
		task84APostActivation, loadTxCount, maxBlockTxs, len(durations), p99, latencyLimit, heights, hashes[0], len(malicious), task84ATimeZones)
}

func buildSignedLoadTxs(t *testing.T, net *networktest.Network, cfg networktest.Config, count int) [][]byte {
	t.Helper()
	validator := net.Validators[0]
	status, err := validator.RPCClient.Status(context.Background())
	require.NoError(t, err)
	timeoutBase := status.SyncInfo.LatestBlockTime.Add(5 * time.Minute)
	const transactionsPerSigner = int(apptypes.DefaultMaxTxPerBlockPerAccount)
	signerCount := (count + transactionsPerSigner - 1) / transactionsPerSigner
	require.LessOrEqual(t, signerCount, len(net.AdditionalAccounts))
	transactions := make([][]byte, count)
	transactionIndex := 0
	for signerIndex := 0; signerIndex < signerCount; signerIndex++ {
		signer := net.AdditionalAccounts[signerIndex]
		name := signer.Name
		address := signer.Address
		accountNumber, _, err := cfg.AccountRetriever.GetAccountNumberSequence(validator.ClientCtx.WithClient(validator.RPCClient), address)
		require.NoError(t, err)
		message := banktypes.NewMsgSend(address, address, sdk.NewCoins(sdk.NewInt64Coin(cfg.BondDenom, 1)))
		baseFactory := tx.Factory{}.
			WithChainID(cfg.ChainID).
			WithKeybase(validator.ClientCtx.Keyring).
			WithTxConfig(cfg.TxConfig).
			WithAccountNumber(accountNumber).
			WithSequence(0).
			WithUnordered(true).
			WithGas(task84ATxGas).
			WithFees("1" + cfg.BondDenom)
		for signerSequence := 0; signerSequence < transactionsPerSigner && transactionIndex < count; signerSequence++ {
			factory := baseFactory.
				WithTimeoutTimestamp(timeoutBase.Add(time.Duration(transactionIndex+1) * time.Nanosecond)).
				WithMemo(fmt.Sprintf("task84a-load-%05d", transactionIndex))
			builder, err := factory.BuildUnsignedTx(message)
			require.NoError(t, err)
			require.NoError(t, tx.Sign(context.Background(), factory, name, builder, true))
			transactions[transactionIndex], err = cfg.TxConfig.TxEncoder()(builder.GetTx())
			require.NoError(t, err)
			transactionIndex++
		}
	}
	return transactions
}

func validateLoadThroughCheckTx(t *testing.T, net *networktest.Network, transactions [][]byte) {
	t.Helper()
	rawApp, err := net.ValidatorApp(0)
	require.NoError(t, err)
	observed, ok := rawApp.(*observedApplication)
	require.True(t, ok)
	for index, txBytes := range transactions {
		result, err := observed.CheckTx(&abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New})
		require.NoError(t, err, "broadcast %d failed", index)
		require.Zero(t, result.Code, "CheckTx %d rejected: %s", index, result.Log)
	}
}

func requireAllValidatorsAtLeast(t *testing.T, ctx context.Context, net *networktest.Network, target int64) {
	t.Helper()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	latest := make([]int64, len(net.Validators))
	for {
		allReached := true
		for index := range net.Validators {
			status, err := net.Validators[index].RPCClient.Status(ctx)
			if err != nil {
				allReached = false
				continue
			}
			latest[index] = status.SyncInfo.LatestBlockHeight
			if latest[index] < target {
				allReached = false
			}
		}
		if allReached {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatalf("validators did not reach height %d before timeout; latest=%v: %v", target, latest, ctx.Err())
		case <-ticker.C:
		}
	}
}

func requireStrictActivation(t *testing.T, net *networktest.Network, activationHeight int64) {
	t.Helper()
	for index := range net.Validators {
		rawApp, err := net.ValidatorApp(index)
		require.NoError(t, err)
		observed, ok := rawApp.(*observedApplication)
		require.True(t, ok)
		validatorApp, ok := observed.Application.(*app.VirtEngineApp)
		require.True(t, ok)
		ctx := validatorApp.NewUncachedContext(false, cmtproto.Header{ChainID: net.Config.ChainID, Height: activationHeight})
		doneHeight, err := validatorApp.Keepers.Cosmos.Upgrade.GetDoneHeight(ctx, utypes.ConsensusAdmissionUpgradeName)
		require.NoError(t, err)
		require.Equal(t, task84AUpgradeHeight, doneHeight, "validator %d is not activated by the executed upgrade", index)
		params, err := validatorApp.Keepers.Cosmos.ConsensusParams.ParamsStore.Get(ctx)
		require.NoError(t, err)
		require.NotNil(t, params.Abci)
		require.Equal(t, activationHeight, params.Abci.VoteExtensionsEnableHeight)
	}
}

func validatorEvidenceAtHeight(t *testing.T, ctx context.Context, net *networktest.Network, height int64) ([]int64, [][]byte) {
	t.Helper()
	heights := make([]int64, len(net.Validators))
	hashes := make([][]byte, len(net.Validators))
	for index := range net.Validators {
		status, err := net.Validators[index].RPCClient.Status(ctx)
		require.NoError(t, err)
		heights[index] = status.SyncInfo.LatestBlockHeight
		results := waitForBlockResults(t, ctx, net.Validators[index], height)
		require.Equal(t, height, results.Height)
		hashes[index] = bytes.Clone(results.AppHash)
		require.NotEmpty(t, hashes[index])
	}
	return heights, hashes
}

func waitForBlockResults(t *testing.T, ctx context.Context, validator *networktest.Validator, height int64) *coretypes.ResultBlockResults {
	t.Helper()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		results, err := validator.RPCClient.BlockResults(ctx, &height)
		if err == nil {
			return results
		}
		select {
		case <-ctx.Done():
			t.Fatalf("validator %d results for height %d unavailable: %v", validator.Index, height, err)
		case <-ticker.C:
		}
	}
}

func assertMaliciousProposalRejected(t *testing.T, observations []processObservation) {
	t.Helper()
	rejected := make(map[int]bool)
	for _, observation := range observations {
		if observation.Status == abci.ResponseProcessProposal_REJECT {
			rejected[observation.Validator] = true
		}
	}
	require.GreaterOrEqual(t, len(rejected), 3, "at least three honest validators must reject the malicious proposal")
}

func strictProcessDurations(observations []processObservation, firstHeight, lastHeight int64) ([]time.Duration, map[int]int) {
	durations := make([]time.Duration, 0, len(observations))
	perValidator := make(map[int]int)
	for _, observation := range observations {
		if observation.Height < firstHeight || observation.Height > lastHeight || observation.Status != abci.ResponseProcessProposal_ACCEPT {
			continue
		}
		durations = append(durations, observation.Duration)
		perValidator[observation.Validator]++
	}
	return durations, perValidator
}

func blockLoadEvidence(t *testing.T, ctx context.Context, validator *networktest.Validator, rawApp servertypes.Application, firstHeight, lastHeight int64) (int, int) {
	t.Helper()
	validatorApp, ok := validatorAppFromApplication(rawApp)
	require.True(t, ok)
	maxTxs := 0
	totalOrdinary := 0
	for height := firstHeight; height <= lastHeight; height++ {
		block, err := validator.RPCClient.Block(ctx, &height)
		require.NoError(t, err)
		txCount := len(block.Block.Txs)
		if txCount > maxTxs {
			maxTxs = txCount
		}
		if txCount > 0 {
			require.True(t, isConsensusSystemTx(validatorApp, block.Block.Txs[0]), "height %d lacks the strict index-zero carrier", height)
			totalOrdinary += txCount - 1
		}
	}
	return maxTxs, totalOrdinary
}

func isConsensusSystemTx(rawApp servertypes.Application, txBytes []byte) bool {
	validatorApp, ok := validatorAppFromApplication(rawApp)
	if !ok {
		return false
	}
	decoded, err := validatorApp.TxConfig().TxDecoder()(txBytes)
	if err != nil {
		return false
	}
	msgs := decoded.GetMsgs()
	if len(msgs) != 1 {
		return false
	}
	_, ok = msgs[0].(*veidv1.MsgSubmitConsensusVerification)
	return ok
}

func validatorAppFromApplication(rawApp servertypes.Application) (*app.VirtEngineApp, bool) {
	if observed, ok := rawApp.(*observedApplication); ok {
		rawApp = observed.Application
	}
	validatorApp, ok := rawApp.(*app.VirtEngineApp)
	return validatorApp, ok
}

func netApp(t *testing.T, net *networktest.Network, index int) servertypes.Application {
	t.Helper()
	rawApp, err := net.ValidatorApp(index)
	require.NoError(t, err)
	return rawApp
}

func percentile99(samples []time.Duration) time.Duration {
	if len(samples) == 0 {
		return 0
	}
	sorted := append([]time.Duration(nil), samples...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	index := (99*len(sorted) + 99) / 100
	return sorted[index-1]
}

func minDuration(left, right time.Duration) time.Duration {
	if left < right {
		return left
	}
	return right
}

func envInt(t *testing.T, name string, fallback int) int {
	t.Helper()
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	require.NoError(t, err, "%s must be an integer", name)
	require.Positive(t, parsed, "%s must be positive", name)
	return parsed
}

var _ = veidv1.MsgSubmitConsensusVerification{}
