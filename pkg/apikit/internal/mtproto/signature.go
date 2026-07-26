package mtproto

import (
	"fmt"

	"mtlogin/pkg/util"
)

func Signature(method, path string, timestamp int64) string {
	message := fmt.Sprintf("%s&%s&%d", method, path, timestamp)
	return util.HmacSha1Base64(message, "HLkPcWmycL57mfJt")
}
