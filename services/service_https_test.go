package services

import (
	"fmt"
	"net"
	"testing"
)

func TestHTTPS_Scan(t *testing.T) {
	var task = Task{
		ip:   "baidu.com",
		port: "443",
	}
	conn, _ := net.Dial("tcp", task.ip+":"+task.port)

	mq := HTTPS{}
	service, banner, err := mq.Scan(conn, task, nil)
	fmt.Println(service)
	fmt.Println(string(banner))
	fmt.Println(err)
}
