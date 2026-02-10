package v1beta3

type ClientOptions struct {
	tclient TxClient // nolint: unused
}

type ClientOption func(*ClientOptions) *ClientOptions
