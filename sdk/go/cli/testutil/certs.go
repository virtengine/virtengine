package testutil

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	sdktest "github.com/cosmos/cosmos-sdk/testutil"

	"github.com/virtengine/virtengine/sdk/go/cli"
)

// TxGenerateServerExec is used for testing create server certificate tx
func TxGenerateServerExec(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetTxCertGenerateServerCmd(), args...)
}

// TxGenerateClientExec is used for testing create client certificate tx
func TxGenerateClientExec(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetTxCertGenerateClientCmd(), args...)
}

// TxPublishServerExec is used for testing create server certificate tx
func TxPublishServerExec(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetTxCertPublishServerCmd(), args...)
}

// TxPublishClientExec is used for testing create client certificate tx
func TxPublishClientExec(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetTxCertPublishClientCmd(), args...)
}

// TxRevokeServerExec is used for testing create server certificate tx
func TxRevokeServerExec(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetTxCertRevokeServerCmd(), args...)
}

// TxRevokeClientExec is used for testing create client certificate tx
func TxRevokeClientExec(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetTxCertsRevokeClientCmd(), args...)
}

// QueryCertificatesExec is used for testing certificates query
func QueryCertificatesExec(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetQueryCertCertificatesCmd(), args...)
}

// QueryCertificateExec is used for testing certificate query
func QueryCertificateExec(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetQueryCertCertificatesCmd(), args...)
}
