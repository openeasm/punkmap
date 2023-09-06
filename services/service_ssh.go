package services

import (
	"bytes"
	"easm_punkmap/common"
	"net"
)

type SSH struct {
}

func (s *SSH) Scan(conn net.Conn, task Task, result *Result) (service string, banner []byte, err error) {
	// no need send data, read data directly
	// read data from conn
	banner, err = common.ReadUntilNewLine(conn)
	if banner != nil {
		if bytes.HasPrefix(banner, []byte("SSH-")) || bytes.HasPrefix(banner, []byte("ssh-")) {
			return "ssh", banner, nil
		} else {
			return "", banner, nil
		}
	} else {
		return "", banner, err
	}
}

func (s *SSH) DefaultPorts() []string {
	return []string{"22", "2222"}
}
func init() {
	RegistScanner(&SSH{})
}
