package services

import (
	"bytes"
	"easm_punkmap/common"
	"errors"
	"net"
)

type RocketMQ struct {
	ServiceScanner
}

func (s *RocketMQ) Scan(conn net.Conn, task Task) (service string, banner []byte, err error) {
	handshake := common.MustBase64Decode("AAAA0AAAALJ7ImNvZGUiOjMxOCwiZXh0RmllbGRzIjp7IkFjY2Vzc0tleSI6InJvY2tldG1xMiIsIlNpZ25hdHVyZSI6ImNHSmpxMUZCTSs0VUJsUnNORE50azBVOW5EMD0ifSwiZmxhZyI6MCwibGFuZ3VhZ2UiOiJKQVZBIiwib3BhcXVlIjowLCJzZXJpYWxpemVUeXBlQ3VycmVudFJQQyI6IkpTT04iLCJ2ZXJzaW9uIjo0MzV9dGhpc19pc19rZXk9dGhpc19pc192YWx1ZQo=")
	write, err := conn.Write(handshake)
	if err != nil {
		return "", nil, err
	}
	if write != len(handshake) {
		return "", nil, errors.New("write error")
	}
	// read data from conn
	banner, err = common.ReadAll(conn)
	if err != nil {
		return "", nil, err
	}
	if bytes.Contains(banner, []byte("serializeTypeCurrentRPC")) {
		return "rocketmq", banner, nil
	} else {
		return "", banner, nil
	}
}
func (s *RocketMQ) DefaultPorts() []string {
	return []string{"9876"}
}
func init() {
	RegistScanner(&RocketMQ{})
}
