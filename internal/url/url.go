package url

import (
	"crypto/sha256"
	"encoding/hex"
)

func Shorten(originalUrl string) string {
	h := sha256.New()
	h.Write([]byte(originalUrl))
	hash := hex.EncodeToString(h.Sum(nil))
	shortURL := hash[:8]
	return shortURL
}
