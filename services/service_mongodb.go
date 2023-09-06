package services

import (
	"bytes"
	"easm_punkmap/common"
	"net"
)

type MongoDB struct {
}

func (s *MongoDB) Scan(conn net.Conn, task Task, result *Result) (service string, banner []byte, err error) {
	// no need send data, read data directly
	// read data from conn
	// write mongodb handshake data, send ping
	_, err = conn.Write([]byte("\x00\x00\x00\x00\x00\x00\x00\x00\x70\x00\x00\x00\x00\x00\x00\x00\xd4\x07\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00admin.$cmd\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01\x00\x00\x00\x10\x70\x69\x6e\x67\x00\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"))
	banner, err = common.ReadAll(conn)
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

func (s *MongoDB) DefaultPorts() []string {
	return []string{"27017"}
}
func init() {
	RegistScanner(&MongoDB{})
}
