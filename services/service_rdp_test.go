package services

import (
	"fmt"
	"net"
	"testing"
)

func TestRDP_Scan(t *testing.T) {
	var task = Task{
		ip:   "155.159.57.105",
		port: "3389",
	}
	conn, _ := net.Dial("tcp", task.ip+":"+task.port)

	mq := RDP{}
	service, banner, err := mq.Scan(conn, task)
	fmt.Println(service)
	fmt.Println(string(banner))
	fmt.Println(err)
}
