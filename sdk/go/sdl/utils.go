package sdl

const (
	valueFalse = "false"
	valueTrue  = "true"
)

// as per yaml following allowed as bool values
func unifyStringAsBool(val string) (string, bool) {
	switch val {
	case valueTrue, "on", "yes":
		return valueTrue, true
	case valueFalse, "off", "no":
		return valueFalse, true
	}

	return "", false
}
