package units

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/virtengine/virtengine/sdk/go/node/types/unit"
)

func TestByteQuantity(t *testing.T) {
	type vtype struct {
		Val ByteQuantity `yaml:"val"`
	}

	tests := []struct {
		text  string
		value uint64
		err   bool
	}{
		{`val: 1`, 1, false},
		{`val: -1`, 0, true},

		{`val: 01`, 0, true},

		{`val: -01`, 0, true},

		{`val: "1M"`, unit.M, false},
		{`val: "-1M"`, 0, true},

		{`val: "0.5M"`, unit.M / 2, false},
		{`val: "-0.5M"`, 0, true},

		{`val: "00.5M"`, 0, true},
		{`val: "-00.5M"`, 0, true},

		{`val: "3M"`, 3 * unit.M, false},
		{`val: "3G"`, 3 * unit.G, false},
		{`val: "3T"`, 3 * unit.T, false},
		{`val: "3P"`, 3 * unit.P, false},
		{`val: "3E"`, 3 * unit.E, false},

		{`val: ""`, 0, true},
	}

	for idx, test := range tests {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			buf := []byte(test.text)
			obj := &vtype{}

			err := yaml.Unmarshal(buf, obj)

			if test.err {
				assert.Error(t, err)
				assert.Equal(t, ByteQuantity(test.value), obj.Val)
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, ByteQuantity(test.value), obj.Val)
		})
	}
}

