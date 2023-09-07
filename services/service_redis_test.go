package services

import (
	"fmt"
	"net"
	"testing"
)

func TestRedis_Scan(t *testing.T) {
	conn, _ := net.Dial("tcp", "121.40.92.218:6379")
	// send redis probe
	conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
	// read response
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		t.Error(err)
	}
	if string(buffer[:n]) != "+PONG\r\n" {
		t.Error("invalid response")
		fmt.Println(string(buffer[:n]))
	} else {
		t.Log("redis probe success")
	}
}

func TestRedis_Scan1(t *testing.T) {
	conn, _ := net.Dial("tcp", "121.40.92.218:6379")
	// send redis probe
	redis := Redis{}
	service, banner, err := redis.Scan(conn, Task{}, &Result{})
	if err != nil {
		t.Error(err)
	} else {
		fmt.Println(service)
		fmt.Println(string(banner))
	}
}
