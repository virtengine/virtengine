package v1

import (
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	sdkerrors "cosmossdk.io/errors"
	"google.golang.org/grpc/codes"
)

const (
	ModuleName                        = "akash"
	AttributeNameRegexpStringWildcard = `^([a-zA-Z][\w\/\.\-]{1,126}[\w\*]?)$`
	AttributeNameRegexpString         = `^([a-zA-Z][\w\/\.\-]{1,126})$`
)

const (
	errAttributesDuplicateKeys uint32 = iota + 1
	errInvalidAttributeKey
	ErrLast
)

var (
	ErrAttributesDuplicateKeys = sdkerrors.RegisterWithGRPCCode(ModuleName, errAttributesDuplicateKeys, codes.InvalidArgument, "attributes cannot have duplicate keys")
	ErrInvalidAttributeKey     = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidAttributeKey, codes.InvalidArgument, "attribute key does not match regexp")
)

var (
	attributeNameRegexpWildcard = regexp.MustCompile(AttributeNameRegexpStringWildcard)
)

/*
Attributes purpose of using this type in favor of Cosmos's sdk.Attribute is
ability to later extend it with operators to support querying on things like
cpu/memory/storage attributes
At this moment type though is same as sdk.Attributes but all akash libraries were
turned to use a new one
*/
type Attributes []Attribute

var _ sort.Interface = (*Attributes)(nil)

type AttributesGroup []Attributes

type AttributeValue interface {
	AsBool() (bool, bool)
	AsString() (string, bool)
}

type attributeValue struct {
	value string
}

func (val attributeValue) AsBool() (bool, bool) {
	if val.value == "" {
		return false, false
	}

	res, err := strconv.ParseBool(val.value)
	if err != nil {
		return false, false
	}

	return res, true
}

func (val attributeValue) AsString() (string, bool) {
	if val.value == "" {
		return "", false
	}

	return val.value, true
}

func (m PlacementRequirements) Dup() PlacementRequirements {
	return PlacementRequirements{
		SignedBy:   m.SignedBy,
		Attributes: m.Attributes.Dup(),
	}
}

func NewStringAttribute(key, val string) Attribute {
	return Attribute{
		Key:   key,
		Value: val,
	}
}

func (m *Attribute) String() string {
	res, _ := yaml.Marshal(m)
	return string(res)
}

func (m *Attribute) Equal(rhs *Attribute) bool {
	return reflect.DeepEqual(m, rhs)
}

func (m Attribute) SubsetOf(rhs Attribute) bool {
	match, _ := filepath.Match(m.Key, rhs.Key)

	return match && (m.Value == rhs.Value)
}

func (attr Attributes) Len() int {
	return len(attr)
}

func (attr Attributes) Swap(i, j int) {
	attr[i], attr[j] = attr[j], attr[i]
}

func (attr Attributes) Less(i, j int) bool {
	return attr[i].Key < attr[j].Key
}

func (attr Attributes) Validate() error {
	return attr.ValidateWithRegex(attributeNameRegexpWildcard)
}

func (attr Attributes) ValidateWithRegex(r *regexp.Regexp) error {
	store := make(map[string]bool)

	for i := range attr {
		if !r.MatchString(attr[i].Key) {
			return ErrInvalidAttributeKey
		}

		if _, ok := store[attr[i].Key]; ok {
			return ErrAttributesDuplicateKeys
		}

		store[attr[i].Key] = true
	}

	return nil
}

func (attr Attributes) Dup() Attributes {
	if attr == nil {
		return nil
	}

	res := make(Attributes, 0, len(attr))

	for _, pair := range attr {
		res = append(res, Attribute{
			Key:   pair.Key,
			Value: pair.Value,
		})
	}

	return res
}

// AttributesSubsetOf check if a is subset of b
// nolint: gofmt
// For example there are two yaml files being converted into these attributes
// example 1: a is subset of b
// ---
// // a
// attributes:
//
//	region:
//	  - us-east-1
//
// ---
// b
// attributes:
//
//	region:
//	  - us-east-1
//	  - us-east-2
//
// example 2: a is not subset of b
// attributes:
//
//	region:
//	  - us-east-1
//
// ---
// b
// attributes:
//
//	region:
//	  - us-east-2
//	  - us-east-3
//
// example 3: a is subset of b
// attributes:
//
//	region:
//	  - us-east-2
//	  - us-east-3
//
// ---
// b
// attributes:
//
//	region:
//	  - us-east-2
func AttributesSubsetOf(a, b Attributes) bool {
loop:
	for _, req := range a {
		for _, attr := range b {
			if req.SubsetOf(attr) {
				continue loop
			}
		}
		return false
	}

	return true
}

func AttributesAnyOf(a, b Attributes) bool {
	for _, req := range a {
		for _, attr := range b {
			if req.SubsetOf(attr) {
				return true
			}
		}
	}

	return false
}

func (attr Attributes) SubsetOf(b Attributes) bool {
	return AttributesSubsetOf(attr, b)
}

func (attr Attributes) AnyOf(b Attributes) bool {
	return AttributesAnyOf(attr, b)
}

func (attr Attributes) Find(glob string) AttributeValue {
	// todo wildcard

	var val attributeValue

	for i := range attr {
		if glob == attr[i].Key {
			val.value = attr[i].Value
			break
		}
	}

	return val
}

func (attr Attributes) Iterate(prefix string, fn func(group, key, value string)) {
	for _, item := range attr {
		if strings.HasPrefix(item.Key, prefix) {
			tokens := strings.SplitAfter(item.Key, "/")
			tokens = tokens[1:]
			fn(tokens[1], tokens[2], item.Value)
		}
	}
}

// GetCapabilitiesGroup
//
// example
// capabilities/storage/1/persistent: true
// capabilities/storage/1/class: io1
// capabilities/storage/2/persistent: false
//
// nolint: gofmt
// returns
//   - - persistent: true
//     class: nvme
//   - - persistent: false
func (attr Attributes) GetCapabilitiesGroup(prefix string) AttributesGroup {
	var res AttributesGroup // nolint:prealloc

	groups := make(map[string]Attributes)

	for _, item := range attr {
		if !strings.HasPrefix(item.Key, "capabilities/"+prefix) {
			continue
		}

		tokens := strings.SplitAfter(strings.TrimPrefix(item.Key, "capabilities/"), "/")
		// skip malformed attributes. really?
		if len(tokens) != 3 {
			continue
		}

		// filter out prefix name
		tokens = tokens[1:]

		group := groups[tokens[0]]
		if group == nil {
			group = Attributes{}
		}

		group = append(group, Attribute{
			Key:   tokens[1],
			Value: item.Value,
		})

		groups[tokens[0]] = group
	}

	for _, group := range groups {
		res = append(res, group)
	}

	return res
}

func (attr Attributes) GetCapabilitiesMap(prefix string) AttributesGroup {
	res := make(AttributesGroup, 0, 1)
	groups := make(Attributes, 0, len(attr))

	for _, item := range attr {
		if !strings.HasPrefix(item.Key, "capabilities/"+prefix) {
			continue
		}

		tokens := strings.Split(strings.TrimPrefix(item.Key, "capabilities/"), "/")
		// skip malformed attributes
		if len(tokens) < 3 {
			continue
		}

		// filter out prefix name
		tokens = tokens[1:]

		var key string
		for i, token := range tokens {
			if i == 0 {
				key = token
			} else {
				key += "/" + token
			}
		}

		groups = append(groups, Attribute{
			Key:   key,
			Value: item.Value,
		})
	}

	res = append(res, groups)

	return res
}

// IN check if given attributes are in attributes group
// AttributesGroup for storage
//   - persistent: true
//     class: beta1
//   - persistent: true
//     class: beta2
//
// that
//   - persistent: true
//     class: beta1
func (attr Attributes) IN(group AttributesGroup) bool {
	for _, group := range group {
		if attr.SubsetOf(group) {
			return true
		}
	}
	return false
}

func (attr Attributes) AnyIN(group AttributesGroup) bool {
	for _, group := range group {
		if attr.AnyOf(group) {
			return true
		}
	}
	return false
}
