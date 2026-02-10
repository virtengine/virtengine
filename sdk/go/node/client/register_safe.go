package client

import (
	"log"

	proto "github.com/cosmos/gogoproto/proto"
)

func safeRegisterFile(path string, descriptor []byte) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("gogoproto: skipped registering %s due to corrupt descriptor: %v", path, r)
		}
	}()

	proto.RegisterFile(path, descriptor)
}
