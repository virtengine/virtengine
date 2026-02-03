package v1

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ProviderID struct {
	Owner   sdk.Address
	Auditor sdk.Address
}

// AuditedProviders is the collection of AuditedProvider
type AuditedProviders []AuditedProvider

// String implements the Stringer interface for a Providers object.
func (obj AuditedProviders) String() string {
	var buf bytes.Buffer

	const sep = "\n\n"

	for _, p := range obj {
		buf.WriteString(p.String())
		buf.WriteString(sep)
	}

	if len(obj) > 0 {
		buf.Truncate(buf.Len() - len(sep))
	}

	return buf.String()
}
