package static

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
)

var StylesCSSVersion string

func init() {
	StylesCSSVersion = hashFile("./web/static/css/output.css")
}

func hashFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return "0"
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:8])
}
