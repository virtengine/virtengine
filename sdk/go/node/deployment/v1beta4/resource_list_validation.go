package v1beta4

import (
	"errors"
)

var (
	ErrNoGroupsPresent = errors.New("validation: no groups present")
	ErrGroupEmptyName  = errors.New("validation: group has empty name")
)

// func ValidateResourceList(rlist GSpec) error {
// 	if rlist.GetName() == "" {
// 		return ErrGroupEmptyName
// 	}
//
// 	units := rlist.GetResources()
//
// 	if count := len(units); count > validationConfig.MaxGroupUnits {
// 		return fmt.Errorf("group %v: too many units (%v > %v)", rlist.GetName(), count, validationConfig.MaxGroupUnits)
// 	}
//
// 	if err := units.Validate(); err != nil {
// 		return fmt.Errorf("group %v: %w", rlist.GetName(), err)
// 	}
//
// 	limits := newLimits()
//
// 	for _, resource := range units {
// 		gLimits, err := validateGroupResource(resource)
// 		if err != nil {
// 			return fmt.Errorf("group %v: %w", rlist.GetName(), err)
// 		}
//
// 		limits.add(gLimits)
// 	}
//
// 	if limits.cpu.GT(sdkmath.NewIntFromUint64(validationConfig.MaxGroupCPU)) || limits.cpu.LTE(sdk.ZeroInt()) {
// 		return fmt.Errorf("group %v: invalid total CPU (%v > %v > %v fails)",
// 			rlist.GetName(), validationConfig.MaxGroupCPU, limits.cpu, 0)
// 	}
//
// 	if !limits.gpu.IsZero() && (limits.gpu.GT(sdkmath.NewIntFromUint64(validationConfig.MaxGroupGPU)) || limits.gpu.LTE(sdk.ZeroInt())) {
// 		return fmt.Errorf("group %v: invalid total GPU (%v > %v > %v fails)",
// 			rlist.GetName(), validationConfig.MaxGroupGPU, limits.gpu, 0)
// 	}
//
// 	if limits.memory.GT(sdkmath.NewIntFromUint64(validationConfig.MaxGroupMemory)) || limits.memory.LTE(sdk.ZeroInt()) {
// 		return fmt.Errorf("group %v: invalid total memory (%v > %v > %v fails)",
// 			rlist.GetName(), validationConfig.MaxGroupMemory, limits.memory, 0)
// 	}
//
// 	for i := range limits.storage {
// 		if limits.storage[i].GT(sdkmath.NewIntFromUint64(validationConfig.MaxGroupStorage)) || limits.storage[i].LTE(sdk.ZeroInt()) {
// 			return fmt.Errorf("group %v: invalid total storage (%v > %v > %v fails)",
// 				rlist.GetName(), validationConfig.MaxGroupStorage, limits.storage, 0)
// 		}
// 	}
//
// 	return nil
// }

// func validateGroupResource(rg GroupResource) (resourceLimits, error) {
// 	limits, err := validateResourceUnits(rg.Resource.Units)
// 	if err != nil {
// 		return resourceLimits{}, err
// 	}
//
// 	if rg.Count > uint32(validationConfig.MaxUnitCount) || rg.Count < uint32(validationConfig.MinUnitCount) {
// 		return resourceLimits{}, fmt.Errorf("error: invalid unit count (%v > %v > %v fails)",
// 			validationConfig.MaxUnitCount, rg.Count, validationConfig.MinUnitCount)
// 	}
//
// 	limits.mul(rg.Count)
//
// 	return limits, nil
// }

// func validateResourceUnits(units types.ResourceUnits) (resourceLimits, error) {
// 	limits := newLimits()
//
// 	val, err := validateCPU(units.CPU)
// 	if err != nil {
// 		return resourceLimits{}, err
// 	}
// 	limits.cpu = limits.cpu.Add(val)
//
// 	val, err = validateGPU(units.GPU)
// 	if err != nil {
// 		return resourceLimits{}, err
// 	}
// 	limits.gpu = limits.gpu.Add(val)
//
// 	val, err = validateMemory(units.Memory)
// 	if err != nil {
// 		return resourceLimits{}, err
// 	}
// 	limits.memory = limits.memory.Add(val)
//
// 	var storage []sdk.Int
// 	storage, err = validateStorage(units.Storage)
// 	if err != nil {
// 		return resourceLimits{}, err
// 	}
//
// 	// fixme this is not actually sum for storage usecase.
// 	// do we really need sum here?
// 	limits.storage = storage
//
// 	return limits, nil
// }
//
// func validateCPU(u *types.CPU) (sdk.Int, error) {
// 	if u == nil {
// 		return sdk.Int{}, fmt.Errorf("error: invalid unit CPU, cannot be nil")
// 	}
// 	if (u.Units.Value() > uint64(validationConfig.MaxUnitCPU)) || (u.Units.Value() < uint64(validationConfig.MinUnitCPU)) {
// 		return sdk.Int{}, fmt.Errorf("error: invalid unit CPU (%v > %v > %v fails)",
// 			validationConfig.MaxUnitCPU, u.Units.Value(), validationConfig.MinUnitCPU)
// 	}
//
// 	if err := u.Attributes.Validate(); err != nil {
// 		return sdk.Int{}, fmt.Errorf("error: invalid CPU attributes: %w", err)
// 	}
//
// 	return u.Units.Val, nil
// }
//
// func validateGPU(u *types.GPU) (sdk.Int, error) {
// 	if u == nil {
// 		return sdk.Int{}, fmt.Errorf("error: invalid unit GPU, cannot be nil")
// 	}
//
// 	if (u.Units.Value() > uint64(validationConfig.MaxUnitGPU)) || (u.Units.Value() < uint64(validationConfig.MinUnitGPU)) {
// 		return sdk.Int{}, fmt.Errorf("error: invalid unit GPU (%v > %v > %v fails)",
// 			validationConfig.MaxUnitGPU, u.Units.Value(), validationConfig.MinUnitGPU)
// 	}
//
// 	if u.Units.Value() == 0 && len(u.Attributes) > 0 {
// 		return sdk.Int{}, fmt.Errorf("error: invalid GPU state. attributes cannot be present if units == 0")
// 	}
//
// 	if err := u.Attributes.Validate(); err != nil {
// 		return sdk.Int{}, fmt.Errorf("error: invalid GPU attributes: %w", err)
// 	}
//
// 	return u.Units.Val, nil
// }
//
// func validateMemory(u *types.Memory) (sdk.Int, error) {
// 	if u == nil {
// 		return sdk.Int{}, fmt.Errorf("error: invalid unit memory, cannot be nil")
// 	}
// 	if (u.Quantity.Value() > validationConfig.MaxUnitMemory) || (u.Quantity.Value() < validationConfig.MinUnitMemory) {
// 		return sdk.Int{}, fmt.Errorf("error: invalid unit memory (%v > %v > %v fails)",
// 			validationConfig.MaxUnitMemory, u.Quantity.Value(), validationConfig.MinUnitMemory)
// 	}
//
// 	if err := u.Attributes.Validate(); err != nil {
// 		return sdk.Int{}, fmt.Errorf("error: invalid Memory attributes: %w", err)
// 	}
//
// 	return u.Quantity.Val, nil
// }
//
// func validateStorage(u types.Volumes) ([]sdk.Int, error) {
// 	if u == nil {
// 		return nil, fmt.Errorf("error: invalid unit storage, cannot be nil")
// 	}
//
// 	storage := make([]sdk.Int, 0, len(u))
//
// 	for i := range u {
// 		if (u[i].Quantity.Value() > validationConfig.MaxUnitStorage) || (u[i].Quantity.Value() < validationConfig.MinUnitStorage) {
// 			return nil, fmt.Errorf("error: invalid unit storage (%v > %v > %v fails)",
// 				validationConfig.MaxUnitStorage, u[i].Quantity.Value(), validationConfig.MinUnitStorage)
// 		}
//
// 		if err := u[i].Attributes.Validate(); err != nil {
// 			return []sdk.Int{}, fmt.Errorf("error: invalid Storage attributes: %w", err)
// 		}
//
// 		storage = append(storage, u[i].Quantity.Val)
// 	}
//
// 	return storage, nil
// }
