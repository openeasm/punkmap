package services

import (
	"fmt"
	"net"
	"testing"
)

func TestHTTP_Scan(t *testing.T) {
	var task = Task{
		ip:   "www.baidu.com",
		port: "80",
	}
	conn, _ := net.Dial("tcp", task.ip+":"+task.port)

	mq := HTTP{}
	service, banner, err := mq.Scan(conn, task)
	fmt.Println(service)
	fmt.Println(string(banner))
	fmt.Println(err)
}
