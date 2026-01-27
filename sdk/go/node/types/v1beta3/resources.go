package v1beta3

import (
	"fmt"
)

type UnitType int

type Unit interface {
	String() string
}

type ResUnit interface {
	Equals(ResUnit) bool
	Add(unit ResUnit) bool
}

type Volumes []Storage

var _ Unit = (*CPU)(nil)
var _ Unit = (*Memory)(nil)
var _ Unit = (*Storage)(nil)
var _ Unit = (*GPU)(nil)

func (m Resources) Validate() error {
	if m.ID == 0 {
		return fmt.Errorf("resources ID must be > 0")
	}

	if m.CPU == nil {
		return fmt.Errorf("CPU must not be nil")
	}

	if m.GPU == nil {
		return fmt.Errorf("GPU must not be nil")
	}

	if m.Memory == nil {
		return fmt.Errorf("memory must not be nil")
	}

	if m.Storage == nil {
		return fmt.Errorf("storage must not be nil")
	}

	if m.Endpoints == nil {
		return fmt.Errorf("endpoints must not be nil")
	}

	return nil
}

func (m Resources) Dup() Resources {
	res := Resources{
		ID:        m.ID,
		CPU:       m.CPU.Dup(),
		GPU:       m.GPU.Dup(),
		Memory:    m.Memory.Dup(),
		Storage:   m.Storage.Dup(),
		Endpoints: m.Endpoints.Dup(),
	}

	return res
}

func (m Resources) In(rhs Resources) bool {
	if !m.CPU.Equal(rhs.CPU) || !m.GPU.Equal(rhs.GPU) ||
		!m.Memory.Equal(rhs.Memory) || !m.Storage.Equal(rhs.Storage) {
		return false
	}

loop:
	for _, ep := range m.Endpoints {
		for _, rep := range rhs.Endpoints {
			if ep.Equal(rep) {
				continue loop
			}
		}

		return false
	}

	return true
}

func (m CPU) Dup() *CPU {
	return &CPU{
		Units:      m.Units.Dup(),
		Attributes: m.Attributes.Dup(),
	}
}

func (m Memory) Dup() *Memory {
	return &Memory{
		Quantity:   m.Quantity.Dup(),
		Attributes: m.Attributes.Dup(),
	}
}

func (m Storage) Dup() *Storage {
	return &Storage{
		Name:       m.Name,
		Quantity:   m.Quantity.Dup(),
		Attributes: m.Attributes.Dup(),
	}
}

func (m GPU) Dup() *GPU {
	return &GPU{
		Units:      m.Units.Dup(),
		Attributes: m.Attributes.Dup(),
	}
}

func (m Volumes) Equal(rhs Volumes) bool {
	for i := range m {
		if !m[i].Equal(rhs[i]) {
			return false
		}
	}

	return true
}

func (m Volumes) Dup() Volumes {
	res := make(Volumes, 0, len(m))

	for _, storage := range m {
		res = append(res, *storage.Dup())
	}

	return res
}

func (m *CPU) EqualUnits(that *CPU) bool {
	if that == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if !m.Units.Equal(&that.Units) {
		return false
	}

	return true
}

func (m *GPU) EqualUnits(that *GPU) bool {
	if that == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if !m.Units.Equal(&that.Units) {
		return false
	}

	return true
}

func (m *Memory) EqualUnits(that *Memory) bool {
	if that == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if !m.Quantity.Equal(&that.Quantity) {
		return false
	}

	return true
}

func (m Volumes) EqualUnits(that Volumes) bool {
	if len(m) != len(that) {
		return false
	}

	for idx, vol := range m {
		if vol.Name != that[idx].Name {
			return false
		}

		if !vol.Quantity.Equal(&that[idx].Quantity) {
			return false
		}

	}

	return true
}
