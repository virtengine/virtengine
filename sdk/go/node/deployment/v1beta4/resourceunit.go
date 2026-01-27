package v1beta4

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/virtengine/virtengine/sdk/go/node/types/resources/v1beta4"
)

// FullPrice method returns full price of resource
func (r *ResourceUnit) FullPrice() sdk.DecCoin {
	return sdk.NewDecCoinFromDec(r.Price.Denom, r.Price.Amount.MulInt64(int64(r.Count)))
}

func (r *ResourceUnit) Dup() ResourceUnit {
	return ResourceUnit{
		Resources: r.Resources.Dup(),
		Count:     r.Count,
		Price:     r.GetPrice(),
	}
}

func (r *ResourceUnit) validate() error {
	// if r.Count > uint32(validationConfig.MaxUnitCount) || r.Count < uint32(validationConfig.MinUnitCount) {
	// 	return fmt.Errorf("error: invalid unit count (%v > %v > %v fails)",
	// 		validationConfig.MaxUnitCount, r.Count, validationConfig.MinUnitCount)
	// }

	if err := validateResources(r.Resources); err != nil {
		return err
	}

	return nil
}

func (r *ResourceUnit) totalResources() resourceLimits {
	limits := newLimits()

	limits.cpu = limits.cpu.Add(r.CPU.Units.Val)
	limits.gpu = limits.gpu.Add(r.GPU.Units.Val)
	limits.memory = limits.memory.Add(r.Memory.Quantity.Val)

	storage := make([]sdkmath.Int, 0, len(r.Storage))

	for _, vol := range r.Storage {
		storage = append(storage, vol.Quantity.Val)
	}

	// fixme this is not actually sum for storage usecase.
	// do we really need sum here?
	limits.storage = storage

	limits.mul(r.Count)

	return limits
}

func (r *ResourceUnit) validatePricing() error {
	if !r.GetPrice().IsValid() {
		return fmt.Errorf("error: invalid price object")
	}

	if r.Price.Amount.GT(sdkmath.LegacyNewDecFromInt(sdkmath.NewIntFromUint64(validationConfig.Unit.Max.Price))) {
		return fmt.Errorf("error: invalid unit price (%v > %v fails)", validationConfig.Unit.Max.Price, r.Price)
	}

	return nil
}

func validateResources(units types.Resources) error {
	if units.ID == 0 {
		return fmt.Errorf("error: invalid resources ID (> 0 fails)")
	}

	if err := validateCPU(units.CPU); err != nil {
		return err
	}

	if err := validateGPU(units.GPU); err != nil {
		return err
	}

	if err := validateMemory(units.Memory); err != nil {
		return err
	}

	if err := validateStorage(units.Storage); err != nil {
		return err
	}

	return nil
}

func validateCPU(u *types.CPU) error {
	if u == nil {
		return fmt.Errorf("error: invalid unit CPU, cannot be nil")
	}

	if (u.Units.Value() > uint64(validationConfig.Unit.Max.CPU)) || (u.Units.Value() < uint64(validationConfig.Unit.Min.CPU)) {
		return fmt.Errorf("error: invalid unit CPU (%v > %v > %v fails)",
			validationConfig.Unit.Max.CPU, u.Units.Value(), validationConfig.Unit.Max.CPU)
	}

	if err := u.Attributes.Validate(); err != nil {
		return fmt.Errorf("error: invalid CPU attributes: %w", err)
	}

	return nil
}

func validateGPU(u *types.GPU) error {
	if u == nil {
		return fmt.Errorf("error: invalid unit GPU, cannot be nil")
	}

	if (u.Units.Value() > uint64(validationConfig.Unit.Max.GPU)) || (u.Units.Value() < uint64(validationConfig.Unit.Min.GPU)) {
		return fmt.Errorf("error: invalid unit GPU (%v > %v > %v fails)",
			validationConfig.Unit.Max.GPU, u.Units.Value(), validationConfig.Unit.Max.GPU)
	}

	if u.Units.Value() == 0 && len(u.Attributes) > 0 {
		return fmt.Errorf("error: invalid GPU state. attributes cannot be present if units == 0")
	}

	if err := u.Attributes.Validate(); err != nil {
		return fmt.Errorf("error: invalid GPU attributes: %w", err)
	}

	return nil
}

func validateMemory(u *types.Memory) error {
	if u == nil {
		return fmt.Errorf("error: invalid unit memory, cannot be nil")
	}
	if (u.Quantity.Value() > validationConfig.Unit.Max.Memory) || (u.Quantity.Value() < validationConfig.Unit.Min.Memory) {
		return fmt.Errorf("error: invalid unit memory (%v > %v > %v fails)",
			validationConfig.Unit.Max.Memory, u.Quantity.Value(), validationConfig.Unit.Max.Memory)
	}

	if err := u.Attributes.Validate(); err != nil {
		return fmt.Errorf("error: invalid Memory attributes: %w", err)
	}

	return nil
}

func validateStorage(u types.Volumes) error {
	if u == nil {
		return fmt.Errorf("error: invalid unit storage, cannot be nil")
	}

	for i := range u {
		if (u[i].Quantity.Value() > validationConfig.Unit.Max.Storage) || (u[i].Quantity.Value() < validationConfig.Unit.Min.Storage) {
			return fmt.Errorf("error: invalid unit storage (%v > %v > %v fails)",
				validationConfig.Unit.Max.Storage, u[i].Quantity.Value(), validationConfig.Unit.Min.Storage)
		}

		if err := u[i].Attributes.Validate(); err != nil {
			return fmt.Errorf("error: invalid Storage attributes: %w", err)
		}
	}

	return nil
}
