package services

import (
	"fmt"
	"net"
	"testing"
)

func TestSocks5_Scan(t *testing.T) {
	conn, _ := net.Dial("tcp", "47.72.44.238:1080")
	mq := Socks5{}
	protocol, banner, _ := mq.Scan(conn, Task{})
	fmt.Println(protocol)
	fmt.Println(string(banner))
}
