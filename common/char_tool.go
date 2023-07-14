package common

import "encoding/base64"

func MustBase64Decode(s string) (data []byte) {
	// decode s to bytes. use base64 package
	data, _ = base64.StdEncoding.DecodeString(s)
	return data
}
