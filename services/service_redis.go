package services

import (
	"easm_punkmap/common"
	"net"
)

type Redis struct {
}

func (s *Redis) Scan(conn net.Conn, task Task, result *Result) (service string, banner []byte, err error) {
	// no need send data, read data directly
	// read data from conn
	_, err = conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
	if err != nil {
		return "", nil, err
	}
	banner, err = common.ReadUntilNewLine(conn)

	if banner != nil {
		if banner[0] == '+' || banner[0] == '-' {
			return "redis", banner, err
		}
	}
	return "", banner, err
}

func (s *Redis) DefaultPorts() []string {
	return []string{"6379"}
}
func init() {
	RegistScanner(&Redis{})
}
