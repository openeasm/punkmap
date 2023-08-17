package common

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
)

func MustBase64Decode(s string) (data []byte) {
	// decode s to bytes. use base64 package
	data, _ = base64.StdEncoding.DecodeString(s)
	return data
}
func MustGzipEncode(data []byte) (gzipData []byte) {
	// encode data to gziped bytes. use gzip package
	a := bytes.NewBuffer(nil)
	w := gzip.NewWriter(a)
	w.Write(data)
	w.Close()
	return a.Bytes()
}
