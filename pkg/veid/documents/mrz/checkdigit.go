package mrz

func calcCheckDigit(value string) int {
	weights := []int{7, 3, 1}
	sum := 0
	for idx, char := range value {
		sum += charValue(char) * weights[idx%len(weights)]
	}
	return sum % 10
}

func digitValue(char byte) int {
	if char < '0' || char > '9' {
		return -1
	}
	return int(char - '0')
}

func checkMatch(value string, expected int) bool {
	if expected < 0 {
		return true
	}
	return calcCheckDigit(value) == expected
}

func charValue(char rune) int {
	switch {
	case char >= '0' && char <= '9':
		return int(char - '0')
	case char >= 'A' && char <= 'Z':
		return int(char-'A') + 10
	case char == '<':
		return 0
	default:
		return 0
	}
}
