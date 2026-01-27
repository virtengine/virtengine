package v1beta4

import (
	sdkmath "cosmossdk.io/math"
)

type resourceLimits struct {
	cpu     sdkmath.Int
	gpu     sdkmath.Int
	memory  sdkmath.Int
	storage []sdkmath.Int
}

func newLimits() resourceLimits {
	return resourceLimits{
		cpu:    sdkmath.ZeroInt(),
		gpu:    sdkmath.ZeroInt(),
		memory: sdkmath.ZeroInt(),
	}
}

func (u *resourceLimits) add(rhs resourceLimits) {
	u.cpu = u.cpu.Add(rhs.cpu)
	u.gpu = u.gpu.Add(rhs.gpu)
	u.memory = u.memory.Add(rhs.memory)

	// u.storage = u.storage.Add(rhs.storage)
}

func (u *resourceLimits) mul(count uint32) {
	u.cpu = u.cpu.MulRaw(int64(count))
	u.gpu = u.gpu.MulRaw(int64(count))
	u.memory = u.memory.MulRaw(int64(count))

	for i := range u.storage {
		u.storage[i] = u.storage[i].MulRaw(int64(count))
	}
}
