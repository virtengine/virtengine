package types

import (
	"context"
	"testing"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	grpc "google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type mockClientConn struct {
	lastMethod string
	lastReq    interface{}
}

func (m *mockClientConn) Invoke(_ context.Context, method string, args, reply interface{}, _ ...grpc.CallOption) error {
	m.lastMethod = method
	m.lastReq = args
	return nil
}

func (m *mockClientConn) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, nil
}

func TestNewQueryClientUsesWorkloadTemplateRPCPaths(t *testing.T) {
	conn := &mockClientConn{}
	client := NewQueryClient(conn)

	_, err := client.WorkloadTemplate(context.Background(), &QueryWorkloadTemplateRequest{TemplateId: "tmpl", Version: "1.0.0"})
	require.NoError(t, err)
	require.Equal(t, "/virtengine.hpc.v1.WorkloadTemplateQuery/WorkloadTemplate", conn.lastMethod)

	_, err = client.SearchWorkloadTemplates(context.Background(), &QuerySearchWorkloadTemplatesRequest{Query: "gpu"})
	require.NoError(t, err)
	require.Equal(t, "/virtengine.hpc.v1.WorkloadTemplateQuery/SearchWorkloadTemplates", conn.lastMethod)
}

func TestWorkloadTemplateQueryDescriptorRegistered(t *testing.T) {
	require.NotNil(t, gogoproto.MessageType("virtengine.hpc.v1.QueryWorkloadTemplateRequest"))
	require.NotNil(t, gogoproto.MessageType("virtengine.hpc.v1.QuerySearchWorkloadTemplatesResponse"))

	desc, err := gogoproto.HybridResolver.FindDescriptorByName(
		protoreflect.FullName("virtengine.hpc.v1.WorkloadTemplateQuery.WorkloadTemplate"),
	)
	require.NoError(t, err)

	methodDesc, ok := desc.(protoreflect.MethodDescriptor)
	require.True(t, ok)
	require.Equal(t, protoreflect.FullName("virtengine.hpc.v1.QueryWorkloadTemplateRequest"), methodDesc.Input().FullName())
	require.Equal(t, protoreflect.FullName("virtengine.hpc.v1.QueryWorkloadTemplateResponse"), methodDesc.Output().FullName())
}
