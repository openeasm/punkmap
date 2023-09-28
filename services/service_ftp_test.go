package services

import (
	"fmt"
	"net"
	"testing"
)

func TestFTP_Scan(t *testing.T) {
	conn, _ := net.Dial("tcp", "188.128.238.82:21")
	mq := FTP{}
	protocol, banner, _ := mq.Scan(conn, Task{}, &Result{})
	fmt.Println(protocol)
	fmt.Println(string(banner))
}
