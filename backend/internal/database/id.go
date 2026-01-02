package database

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"time"
)

var idCounter uint64

// GenerateID creates a unique ID combining timestamp and random bytes
func GenerateID() string {
	timestamp := time.Now().UnixNano()
	counter := atomic.AddUint64(&idCounter, 1)

	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)

	return fmt.Sprintf("%x%04x%s", timestamp, counter&0xFFFF, hex.EncodeToString(randomBytes))
}
