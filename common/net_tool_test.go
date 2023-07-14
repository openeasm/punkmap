package common

import (
	"fmt"
	"net"
	"testing"
)

func TestReadAll(t *testing.T) {
	conn, _ := net.Dial("tcp", "dl01.imfht.com:80")
	conn.Write([]byte("GET / HTTP/1.1\r\nHost: dl01.imfht.com\r\n\r\n"))
	data, _ := ReadAll(conn)
	fmt.Println(string(data))
}

func TestReadAll_22(t *testing.T) {
	conn, _ := net.Dial("tcp", "dl01.imfht.com:22")

	data, _ := ReadAll(conn)
	fmt.Println(string(data))
}
