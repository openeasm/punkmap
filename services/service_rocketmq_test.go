package services

import (
	"fmt"
	"net"
	"testing"
)

func TestRocketMQ_Scan(t *testing.T) {
	conn, _ := net.Dial("tcp", "47.254.127.216:9876")
	mq := RocketMQ{}
	fmt.Println(mq.Scan(conn, Task{}))
}
