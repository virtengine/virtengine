package nfc

import "context"

type AccessProtocol string

const (
	ProtocolBAC  AccessProtocol = "bac"
	ProtocolPACE AccessProtocol = "pace"
	ProtocolEAC  AccessProtocol = "eac"
)

type ChipData struct {
	DG1           []byte
	DG2           []byte
	DG11          []byte
	DG12          []byte
	SOD           []byte
	PassiveValid  bool
	ActiveValid   bool
	ChipAuthValid bool
	ProtocolUsed  AccessProtocol
}

type Reader interface {
	Read(ctx context.Context, mrzKey string) (*ChipData, error)
}
