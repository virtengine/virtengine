package sdl

import (
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"

	types "github.com/virtengine/virtengine/sdk/go/node/types/attributes/v1"
)

type v2CPUAttributes types.Attributes

type v2ResourceCPU struct {
	Units      cpuQuantity     `yaml:"units"`
	Attributes v2CPUAttributes `yaml:"attributes,omitempty"`
}

func (sdl *v2CPUAttributes) UnmarshalYAML(node *yaml.Node) error {
	var attr v2CPUAttributes

	for i := 0; i+1 < len(node.Content); i += 2 {
		var value string
		switch node.Content[i].Value {
		case "arch":
			if err := node.Content[i+1].Decode(&value); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported cpu attribute \"%s\"", node.Content[i].Value)
		}

		attr = append(attr, types.Attribute{
			Key:   node.Content[i].Value,
			Value: value,
		})
	}

	sort.Sort(types.Attributes(attr))

	*sdl = attr

	return nil
}

