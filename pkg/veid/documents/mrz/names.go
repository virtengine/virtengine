package mrz

import "strings"

func ParseName(raw string) (string, string) {
	parts := strings.SplitN(raw, "<<", 2)
	surname := Clean(parts[0])
	given := ""
	if len(parts) > 1 {
		given = Clean(parts[1])
	}
	return surname, given
}

func Clean(value string) string {
	value = strings.ReplaceAll(value, "<", " ")
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return strings.Join(strings.Fields(value), " ")
}
