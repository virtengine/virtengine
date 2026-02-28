package sdkutil

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	pproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/protoadapt"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/gogoproto/proto"
)

type customSigner struct {
	msgType protoreflect.FullName
	field   string
	signer  string
}

var (
	signersLock   sync.RWMutex
	sealed        chan struct{}
	customSigners []customSigner
)

func init() {
	sealed = make(chan struct{})
}

func RegisterCustomSignerField(msg proto.Message, field string, signer string) {
	defer signersLock.Unlock()
	signersLock.Lock()

	select {
	case <-sealed:
		panic("custom signers config has been sealed")
	default:
	}

	msgType := pproto.MessageName(protoadapt.MessageV2Of(msg))

	for _, m := range customSigners {
		if m.msgType == msgType {
			panic(fmt.Sprintf("custom signer for msg \"%s\", has already been registered", msgType.Name()))
		}
	}

	customSigners = append(customSigners, customSigner{
		msgType: msgType,
		field:   field,
		signer:  signer,
	})
}

type CodecOptions struct {
	AccAddressPrefix string
	ValAddressPrefix string
	Options          signing.Options
}

func NewCodecOptions() *CodecOptions {
	return &CodecOptions{
		AccAddressPrefix: Bech32PrefixAccAddr,
		ValAddressPrefix: Bech32PrefixValAddr,
		Options:          NewSigningOptions(),
	}
}

// NewInterfaceRegistry returns a new InterfaceRegistry with the given options.
func (o CodecOptions) NewInterfaceRegistry() codectypes.InterfaceRegistry {
	ir, err := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles:     proto.HybridResolver,
		SigningOptions: o.Options,
	})
	if err != nil {
		panic(err)
	}

	return ir
}

// NewCodec returns a new codec with the given options.
func (o CodecOptions) NewCodec() *codec.ProtoCodec {
	return codec.NewProtoCodec(o.NewInterfaceRegistry())
}

func NewSigningOptions() signing.Options {
	so := signing.Options{
		FileResolver:          nil,
		TypeResolver:          nil,
		AddressCodec:          address.NewBech32Codec(Bech32PrefixAccAddr),
		ValidatorAddressCodec: address.NewBech32Codec(Bech32PrefixValAddr),
		CustomGetSigners:      nil,
		MaxRecursionDepth:     0,
	}

	buildCustomGetSigners(&so)

	return so
}

func BuildCustomSigners() []signing.CustomGetSigner {
	so := NewSigningOptions()
	return buildCustomGetSigners(&so)
}

func getSignerFromID(options *signing.Options, field string, signer string) func(msgIn pproto.Message) ([][]byte, error) {
	return func(msgIn pproto.Message) ([][]byte, error) {
		msg := msgIn.ProtoReflect()
		idDesc := msg.Descriptor().Fields().ByName(protoreflect.Name(field))
		if idDesc != nil {
			if signer == "" {
				if idDesc.Kind() != protoreflect.StringKind {
					return nil, fmt.Errorf("field %s in %s is not a string signer field", field, pproto.MessageName(msgIn))
				}

				b32 := msg.Get(idDesc).String()
				addr, err := options.AddressCodec.StringToBytes(b32)
				if err != nil {
					return nil, fmt.Errorf("error decoding %s address %q: %w", field, b32, err)
				}

				return [][]byte{addr}, nil
			}

			id := msg.Get(idDesc).Message()
			fieldDesc := id.Descriptor().Fields().ByName(protoreflect.Name(signer))
			if fieldDesc == nil {
				return nil, fmt.Errorf("no %s.%s field found in %s", field, signer, pproto.MessageName(msgIn))
			}

			b32 := id.Get(fieldDesc).Interface().(string)
			addr, err := options.AddressCodec.StringToBytes(b32)
			if err != nil {
				return nil, fmt.Errorf("error decoding %s.%s address %q: %w", field, signer, b32, err)
			}

			return [][]byte{addr}, nil
		}

		b32, err := signerFieldFromStruct(msgIn, field, signer)
		if err != nil {
			return nil, err
		}
		addr, err := options.AddressCodec.StringToBytes(b32)
		if err != nil {
			if signer == "" {
				return nil, fmt.Errorf("error decoding %s address %q: %w", field, b32, err)
			}
			return nil, fmt.Errorf("error decoding %s.%s address %q: %w", field, signer, b32, err)
		}

		return [][]byte{addr}, nil
	}
}

func signerFieldFromStruct(msgIn pproto.Message, field string, signer string) (string, error) {
	candidates := make([]reflect.Value, 0, 2)

	if v1 := protoadapt.MessageV1Of(msgIn); v1 != nil {
		candidates = append(candidates, reflect.ValueOf(v1))
	}

	if mt := proto.MessageType(string(pproto.MessageName(msgIn))); mt != nil && mt.Kind() == reflect.Ptr {
		if fresh, ok := reflect.New(mt.Elem()).Interface().(proto.Message); ok {
			if bz, err := pproto.Marshal(msgIn); err == nil {
				if err := proto.Unmarshal(bz, fresh); err == nil {
					candidates = append(candidates, reflect.ValueOf(fresh))
				}
			}
		}
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("nil message value for %s", pproto.MessageName(msgIn))
	}

	var fieldValue reflect.Value
	var err error
	for _, value := range candidates {
		fieldValue, err = lookupProtoStructField(value, field)
		if err == nil {
			break
		}
	}
	if err != nil {
		return "", fmt.Errorf("no \"%s\" field found in %s", field, pproto.MessageName(msgIn))
	}

	if signer == "" {
		if fieldValue.Kind() != reflect.String {
			return "", fmt.Errorf("field %s in %s is not a string signer field", field, pproto.MessageName(msgIn))
		}
		return fieldValue.String(), nil
	}

	if fieldValue.Kind() == reflect.Ptr {
		if fieldValue.IsNil() {
			return "", fmt.Errorf("field %s in %s is nil", field, pproto.MessageName(msgIn))
		}
		fieldValue = fieldValue.Elem()
	}
	if fieldValue.Kind() != reflect.Struct {
		return "", fmt.Errorf("field %s in %s is not a message signer container", field, pproto.MessageName(msgIn))
	}

	signerValue, err := lookupProtoStructField(fieldValue, signer)
	if err != nil {
		return "", fmt.Errorf("no %s.%s field found in %s", field, signer, pproto.MessageName(msgIn))
	}
	if signerValue.Kind() != reflect.String {
		return "", fmt.Errorf("field %s.%s in %s is not a string signer field", field, signer, pproto.MessageName(msgIn))
	}

	return signerValue.String(), nil
}

func lookupProtoStructField(value reflect.Value, protoName string) (reflect.Value, error) {
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return reflect.Value{}, fmt.Errorf("nil pointer")
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("not a struct")
	}

	valueType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		structField := valueType.Field(i)
		tag := structField.Tag.Get("protobuf")
		if tag != "" {
			if name, ok := protoFieldNameFromTag(tag); ok && name == protoName {
				return value.Field(i), nil
			}
		}
	}

	return reflect.Value{}, fmt.Errorf("field not found")
}

func protoFieldNameFromTag(tag string) (string, bool) {
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		if strings.HasPrefix(part, "name=") {
			return strings.TrimPrefix(part, "name="), true
		}
	}
	return "", false
}

func buildCustomGetSigners(options *signing.Options) []signing.CustomGetSigner {
	select {
	case <-sealed:
	default:
		signersLock.Lock()
		close(sealed)
		signersLock.Unlock()
	}

	signers := make([]signing.CustomGetSigner, 0, len(customSigners))
	for _, s := range customSigners {
		signers = append(signers, signing.CustomGetSigner{
			MsgType: s.msgType,
			Fn:      getSignerFromID(options, s.field, s.signer),
		})

	}

	for _, signer := range signers {
		options.DefineCustomGetSigners(signer.MsgType, signer.Fn)
	}

	return signers
}
