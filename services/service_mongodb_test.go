package services

import (
	"fmt"
	"net"
	"testing"
)

func TestMongoDB_Scan(t *testing.T) {
	var task = Task{
		ip:   "47.112.161.97",
		port: "27017",
	}
	conn, _ := net.Dial("tcp", task.ip+":"+task.port)

	mq := MongoDB{}
	service, banner, err := mq.Scan(conn, task)
	fmt.Println(service)
	fmt.Println(string(banner))
	fmt.Println(err)
}
