package sms

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/virtengine/virtengine/pkg/security"
	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type smsChainBackend interface {
	SubmitSMSVerificationProof(ctx context.Context, msg *veidtypes.MsgSubmitSMSVerificationProof) error
	QuerySMSVerification(ctx context.Context, req *veidv1.QuerySMSVerificationRequest) (*veidv1.QuerySMSVerificationResponse, error)
	Close() error
}

type grpcSMSChainBackend struct {
	conn        *grpc.ClientConn
	msgClient   veidv1.MsgClient
	queryClient veidv1.QueryClient
	timeout     time.Duration
}

func newGRPCSMSChainBackend(config ChainIntegrationConfig) (*grpcSMSChainBackend, error) {
	target := resolvedChainGRPCEndpoint(config)
	if target == "" {
		return nil, fmt.Errorf("gRPC endpoint is required")
	}

	dialOptions := []grpc.DialOption{grpc.WithTransportCredentials(resolveChainTransportCredentials(config))}
	conn, err := grpc.NewClient(target, dialOptions...)
	if err != nil {
		return nil, fmt.Errorf("connect SMS chain backend: %w", err)
	}

	timeout := config.RequestTimeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}

	return &grpcSMSChainBackend{
		conn:        conn,
		msgClient:   veidv1.NewMsgClient(conn),
		queryClient: veidv1.NewQueryClient(conn),
		timeout:     timeout,
	}, nil
}

func resolvedChainGRPCEndpoint(config ChainIntegrationConfig) string {
	target := strings.TrimSpace(config.GRPCEndpoint)
	if target == "" {
		target = strings.TrimSpace(config.NodeEndpoint)
	}
	target = strings.TrimPrefix(target, "https://")
	target = strings.TrimPrefix(target, "http://")
	return target
}

func resolveChainTransportCredentials(config ChainIntegrationConfig) credentials.TransportCredentials {
	if config.AllowInsecureTransport {
		return insecure.NewCredentials()
	}
	return credentials.NewTLS(security.SecureTLSConfig())
}

func (b *grpcSMSChainBackend) SubmitSMSVerificationProof(ctx context.Context, msg *veidtypes.MsgSubmitSMSVerificationProof) error {
	reqCtx, cancel := withChainTimeout(ctx, b.timeout)
	defer cancel()

	_, err := b.msgClient.SubmitSMSVerificationProof(reqCtx, msg)
	return err
}

func (b *grpcSMSChainBackend) SubmitSSOVerificationProof(ctx context.Context, msg *veidtypes.MsgSubmitSSOVerificationProof) error {
	return fmt.Errorf("SSO proof submission is not supported by the SMS chain backend")
}

func (b *grpcSMSChainBackend) SubmitEmailVerificationProof(ctx context.Context, msg *veidtypes.MsgSubmitEmailVerificationProof) error {
	return fmt.Errorf("email proof submission is not supported by the SMS chain backend")
}

func (b *grpcSMSChainBackend) SubmitSocialMediaScope(ctx context.Context, msg *veidtypes.MsgSubmitSocialMediaScope) error {
	return fmt.Errorf("social scope submission is not supported by the SMS chain backend")
}

func (b *grpcSMSChainBackend) QuerySMSVerification(ctx context.Context, req *veidv1.QuerySMSVerificationRequest) (*veidv1.QuerySMSVerificationResponse, error) {
	reqCtx, cancel := withChainTimeout(ctx, b.timeout)
	defer cancel()

	return b.queryClient.SMSVerification(reqCtx, req)
}

func (b *grpcSMSChainBackend) Close() error {
	if b.conn == nil {
		return nil
	}
	return b.conn.Close()
}

func withChainTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, timeout)
}
