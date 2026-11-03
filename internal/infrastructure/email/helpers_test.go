package email_test

import (
	"encoding/base64"
	"strings"
)

func decodeBase64(b []byte) ([]byte, error) {
	return base64.StdEncoding.DecodeString(strings.NewReplacer("\r", "", "\n", "").Replace(string(b)))
}
