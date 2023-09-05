package common

import (
	"fmt"
	"net"
	"testing"
)

func TestHeaderParser(t *testing.T) {
	conn, _ := net.Dial("tcp", "dl01.imfht.com:80")
	conn.Write([]byte("GET / HTTP/1.1\r\nHost: dl01.imfht.com\r\n\r\n"))
	data, _ := ReadAll(conn)
	//# jsonify it
	resp := HTTPParser(data)
	fmt.Println(resp)
}
