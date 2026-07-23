package provider_daemon

import (
	"context"
	"net"
	"path/filepath"
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
)

func TestSubmitFiatConversionObservationUsesDurableMutationKind(t *testing.T) {
	chain := newMutationChainFake()
	submitter, _ := newMutationSubmitterForTest(t, chain, filepath.Join(t.TempDir(), "queue.json"))
	msg := validRegistryMessages(submitter.cfg.ProviderAddress)[MutationSettlementFiatObservation].(*settlementv1.MsgRecordFiatConversionObservation)
	result, err := submitter.SubmitFiatConversionObservation(context.Background(), msg)
	require.NoError(t, err)
	require.True(t, result.Final)
	stored, err := submitter.store.Get(context.Background(), result.ID)
	require.NoError(t, err)
	require.Equal(t, MutationSettlementFiatObservation, stored.Kind)
}

type fiatReconciliationQueryServer struct {
	settlementv1.UnimplementedQueryServer
	conversion    *settlementv1.FiatConversionRecord
	params        settlementv1.Params
	financialCase *settlementv1.FinancialCase
}

func (s fiatReconciliationQueryServer) FiatConversion(context.Context, *settlementv1.QueryFiatConversionRequest) (*settlementv1.QueryFiatConversionResponse, error) {
	return &settlementv1.QueryFiatConversionResponse{Conversion: s.conversion}, nil
}
func (s fiatReconciliationQueryServer) Params(context.Context, *settlementv1.QueryParamsRequest) (*settlementv1.QueryParamsResponse, error) {
	return &settlementv1.QueryParamsResponse{Params: s.params}, nil
}
func (s fiatReconciliationQueryServer) FinancialCaseBySubject(context.Context, *settlementv1.QueryFinancialCaseBySubjectRequest) (*settlementv1.QueryFinancialCaseBySubjectResponse, error) {
	return &settlementv1.QueryFinancialCaseBySubjectResponse{FinancialCase: s.financialCase}, nil
}

func TestRPCFiatExecutionAuthorizationIncludesCanonicalFinancialHold(t *testing.T) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	handler := &fiatReconciliationQueryServer{
		conversion:    &settlementv1.FiatConversionRecord{ConversionId: "conversion-hold", SettlementId: "settlement-hold", InvoiceId: "invoice-hold", OrderId: "order-hold"},
		params:        settlementv1.Params{FiatConversionEnabled: true},
		financialCase: &settlementv1.FinancialCase{CaseId: "case-hold", Status: settlementv1.FinancialCaseStatus_FINANCIAL_CASE_STATUS_OPEN, ActiveHoldCount: 2},
	}
	settlementv1.RegisterQueryServer(server, handler)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)
	conn, err := grpc.NewClient("passthrough:///bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	query := &RPCFiatConversionQuery{client: settlementv1.NewQueryClient(conn)}
	authorization, err := query.ExecutionAuthorization(context.Background(), handler.conversion.ConversionId)
	require.NoError(t, err)
	require.True(t, authorization.Params.FiatConversionEnabled)
	require.Equal(t, "case-hold", authorization.ActiveCaseID)
	require.Equal(t, uint32(2), authorization.ActiveHoldCount)

	handler.financialCase = &settlementv1.FinancialCase{CaseId: "case-malformed", Status: settlementv1.FinancialCaseStatus_FINANCIAL_CASE_STATUS_OPEN}
	_, err = query.ExecutionAuthorization(context.Background(), handler.conversion.ConversionId)
	require.ErrorIs(t, err, ErrFiatConversionQueryUnavailable)
}

func TestProviderMutationFiatObservationLogicalReconciliation(t *testing.T) {
	msg := &settlementv1.MsgRecordFiatConversionObservation{ConversionId: "conversion-1", ObservationSequence: 4}
	bz, err := proto.Marshal(msg)
	require.NoError(t, err)
	digest := observationMessageDigest(msg)
	require.NotEmpty(t, bz)

	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	settlementv1.RegisterQueryServer(server, &fiatReconciliationQueryServer{conversion: &settlementv1.FiatConversionRecord{ConversionId: msg.ConversionId, ObservationSequence: msg.ObservationSequence, LastObservationDigest: digest, Observations: []settlementv1.FiatConversionObservation{{Sequence: msg.ObservationSequence, ObservationDigest: digest}}}})
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)
	conn, err := grpc.NewClient("passthrough:///bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	chain := &rpcProviderMutationChain{settlementQuery: settlementv1.NewQueryClient(conn)}
	result, err := chain.ReconcileMutation(context.Background(), &ProviderMutationEnvelope{}, msg)
	require.NoError(t, err)
	require.False(t, result.Committed)
	require.False(t, result.Conflicted)
	require.Empty(t, result.TxHash)
	require.Zero(t, result.Height)
	require.Empty(t, result.BlockHash)

	chain.settlementQuery = settlementv1.NewQueryClient(conn)
	msg.ObservationSequence = 3
	result, err = chain.ReconcileMutation(context.Background(), &ProviderMutationEnvelope{}, msg)
	require.NoError(t, err)
	require.False(t, result.Committed)
	require.True(t, result.Conflicted)
}
