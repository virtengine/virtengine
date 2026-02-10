package participant

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/argon2"
)

type EntropySource struct {
	HardwareRandom []byte
	UserInput      []byte
	SystemNoise    []byte
	Timestamp      time.Time
}

func CollectEntropy(ctx context.Context, duration time.Duration, requireUserInput bool) (*EntropySource, error) {
	hw := make([]byte, 64)
	if _, err := rand.Read(hw); err != nil {
		return nil, err
	}

	var userInput []byte
	if requireUserInput {
		fmt.Println("Please type random characters. Press Enter when finished:")
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadBytes('\n')
		userInput = line
	}

	systemNoise := gatherSystemNoise()
	return &EntropySource{
		HardwareRandom: hw,
		UserInput:      userInput,
		SystemNoise:    systemNoise,
		Timestamp:      time.Now().UTC(),
	}, nil
}

func (e *EntropySource) DeriveSeed() []byte {
	combined := append([]byte{}, e.HardwareRandom...)
	combined = append(combined, e.UserInput...)
	combined = append(combined, e.SystemNoise...)
	combined = append(combined, []byte(e.Timestamp.Format(time.RFC3339Nano))...)
	return argon2.IDKey(combined, []byte("virtengine-trusted-setup"), 1, 64*1024, 4, 64)
}

func gatherSystemNoise() []byte {
	noise := make([]byte, 64)
	_, _ = rand.Read(noise)
	return noise
}
