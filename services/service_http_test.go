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

func BenchmarkFib(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var task = Task{
			ip:   "dl01.imfht.com",
			port: "80",
		}
		conn, _ := net.Dial("tcp", task.ip+":"+task.port)

		mq := HTTP{}
		service, _, _ := mq.Scan(conn, task)
		fmt.Println(service)
	}
}
