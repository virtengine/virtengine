package privileged

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
)

func writeString(output *bytes.Buffer, value string) error {
	length := len(value)
	if length > int(^uint32(0)) {
		return fmt.Errorf("canonical string exceeds uint32 length")
	}
	if err := binary.Write(output, binary.BigEndian, uint32(length)); err != nil {
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

func invalidExactValue(value string) bool {
	trimmed := strings.TrimSpace(value)
	return trimmed == "" || trimmed == "*" || strings.ContainsAny(trimmed, "?[]")
}

func sortedScope(scope []ScopeAtom) []ScopeAtom {
	result := append([]ScopeAtom(nil), scope...)
	sort.Slice(result, func(i, j int) bool { return result[i].key() < result[j].key() })
	return result
}
