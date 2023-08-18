package services

import (
	"easm_punkmap/common"
	"net"
)

type Socks5 struct {
}

func (s *Socks5) Scan(conn net.Conn, task Task) (service string, banner []byte, err error) {
	// send socks5 auth
	// Greeting from Client
	_, err = conn.Write([]byte{0x05, 0x01, 0x00})
	if err != nil {
		return "", nil, err
	}
	banner, err = common.ReadUntilNBytes(conn, 2)

	if banner != nil {
		if banner[0] == 0x05 {
			if banner[1] == 0x00 {
				// no need auth
				// banner name add noauth
				banner = []byte("socks5 noauth (\x05\x00) ")
				return "socks5", banner, err
			} else {
				return "socks5", banner, err
			}
		}
	}
	return "", banner, err
}

func (s *Socks5) DefaultPorts() []string {
	return []string{"1080"}
}
func init() {
	RegistScanner(&Socks5{})
}
