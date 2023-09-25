package services

import (
	"bytes"
	"easm_punkmap/common"
	"net"
)

type FTP struct {
}

func (s *FTP) Scan(conn net.Conn, task Task, result *Result) (service string, banner []byte, err error) {
	// no need send data, read data directly
	// read data from conn
	banner, err = common.ReadUntilNewLine(conn)
	if banner != nil {
		if bytes.HasPrefix(banner, []byte("220")) {
			return "ftp", banner, nil
		} else {
			return "", banner, nil
		}
	} else {
		return "", banner, err
	}
}

func (s *FTP) DefaultPorts() []string {
	return []string{"21"}
}
func init() {
	RegistScanner(&FTP{})
}
