package security_monitoring

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// ID generation helpers

func generateEventID() string {
	return generateID("evt")
}

func generateAlertID() string {
	return generateID("alt")
}

func generateIncidentID() string {
	return generateID("inc")
}

func generateExecutionID() string {
	return generateID("exe")
}

func generateID(prefix string) string {
	data := fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return prefix + "-" + hex.EncodeToString(hash[:8])
}
