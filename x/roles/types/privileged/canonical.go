package privileged

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
)

func writeString(output *bytes.Buffer, value string) error {
	if len(value) > int(^uint32(0)) {
		return fmt.Errorf("canonical string exceeds uint32 length")
	}
	if err := binary.Write(output, binary.BigEndian, uint32(len(value))); err != nil {
		return err
	}
	_, err := output.WriteString(value)
	return err
}

func writeUint64(output *bytes.Buffer, value uint64) {
	_ = binary.Write(output, binary.BigEndian, value)
}

func writeInt64(output *bytes.Buffer, value int64) {
	_ = binary.Write(output, binary.BigEndian, value)
}

func digestStrings(domain string, values ...string) ([32]byte, error) {
	var output bytes.Buffer
	if err := writeString(&output, domain); err != nil {
		return [32]byte{}, err
	}
	for _, value := range values {
		if err := writeString(&output, value); err != nil {
			return [32]byte{}, err
		}
	}
	return sha256.Sum256(output.Bytes()), nil
}

func invalidExactValue(value string) bool {
	trimmed := strings.TrimSpace(value)
	return trimmed == "" || trimmed == "*" || strings.ContainsAny(trimmed, "?[]")
}

func sortedScope(scope []ScopeAtom) []ScopeAtom {
	result := append([]ScopeAtom(nil), scope...)
	sort.Slice(result, func(i, j int) bool { return result[i].key() < result[j].key() })
	return result
}
