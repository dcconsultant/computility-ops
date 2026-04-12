package mysql

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func psaHash(v string) string {
	sum := sha256.Sum256([]byte(v))
	return hex.EncodeToString(sum[:])
}

func nullPSAHash(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return psaHash(v)
}
