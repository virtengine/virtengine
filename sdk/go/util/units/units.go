package units

import (
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/virtengine/virtengine/sdk/go/node/types/unit"
)

type ByteQuantity uint64
type MemoryQuantity uint64

var (
	ErrNegativeValue = fmt.Errorf("units: negative value not allowed")
	ErrOverflow      = fmt.Errorf("units: overflow")
	ErrInvalidFormat = fmt.Errorf("units: invalid format")
	ErrInvalidSuffix = fmt.Errorf("units: invalid suffix")
)

var (
	regexpUnits = regexp.MustCompile(`^([-+]?(([1-9]\d*|0)?[.]\d*|[1-9]\d*))(k|[KMGTPE]i?)?$`)
)

var (
	unitSuffixes = map[string]uint64{
		"":   0,
		"k":  unit.K,
		"Ki": unit.Ki,
		"M":  unit.M,
		"Mi": unit.Mi,
		"G":  unit.G,
		"Gi": unit.Gi,
		"T":  unit.T,
		"Ti": unit.Ti,
		"P":  unit.P,
		"Pi": unit.Pi,
		"E":  unit.E,
		"Ei": unit.Ei,
	}
)

func (u *ByteQuantity) UnmarshalYAML(node *yaml.Node) error {
	val, err := ByteQuantityFromString(node.Value)
	if err != nil {
		return err
	}

	*u = val

	return nil
}

func ToStringWithSuffix(val uint64, suffix string) string {
	unit, exists := unitSuffixes[suffix]

	if exists {
		val /= unit
	} else {
		suffix = ""
	}

	return fmt.Sprintf("%d%s", val, suffix)
}

func (u *ByteQuantity) StringWithSuffix(suffix string) string {
	return ToStringWithSuffix(uint64(*u), suffix)
}

func (u *MemoryQuantity) UnmarshalYAML(node *yaml.Node) error {
	val, err := MemoryQuantityFromString(node.Value)
	if err != nil {
		return err
	}

	*u = val

	return nil
}

func (u *MemoryQuantity) StringWithSuffix(suffix string) string {
	return ToStringWithSuffix(uint64(*u), suffix)
}

func ByteQuantityFromString(sval string) (ByteQuantity, error) {
	val, err := ParseWithSuffix(sval)
	if err != nil {
		return 0, err
	}

	return ByteQuantity(val), nil
}

func MemoryQuantityFromString(sval string) (MemoryQuantity, error) {
	if !strings.HasSuffix(sval, "i") {
		return 0, ErrInvalidSuffix
	}

	val, err := ParseWithSuffix(sval)
	if err != nil {
		return 0, err
	}

	return MemoryQuantity(val), nil
}

func ParseWithSuffix(sval string) (uint64, error) { // strings.SplitAfter()
	match := regexpUnits.FindAllStringSubmatch(sval, -1)
	if len(match) != 1 {
		return 0, ErrInvalidFormat
	}

	tokens := match[0]
	if len(tokens) != 5 {
		return 0, ErrInvalidFormat
	}

	num := tokens[1]
	suffixStr := tokens[4]

	suffix, valid := unitSuffixes[suffixStr]

	if !valid {
		return 0, ErrInvalidSuffix
	}

	val, _, err := new(big.Float).Parse(num, 10)
	if err != nil {
		return 0, ErrInvalidFormat
	}

	if suffix > 0 {
		units := new(big.Float).SetUint64(suffix)
		val = val.Mul(val, units)
	}

	res, acc := val.Uint64()
	if res == 0 && acc == big.Above {
		return 0, ErrNegativeValue
	} else if res == math.MaxUint64 && acc == big.Below {
		return 0, ErrOverflow
	}

	return res, nil
}
