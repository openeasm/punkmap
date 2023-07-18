package services

import (
	"crypto/tls"
	"easm_punkmap/common"
	"fmt"
	"net"
)

type HTTPS struct {
}

func (s *HTTPS) Scan(conn net.Conn, task Task) (service string, banner []byte, err error) {

	// convert conn to tls
	tlsConn := tls.Client(conn, &tls.Config{InsecureSkipVerify: true})
	httpReq := fmt.Sprintf("GET / HTTP/1.1\r\nUser-Agent: PunkMap (https://github.com/openeasm/punkmap)\r\nHost: %s\r\nConnection: close\r\nAccept: */*\r\n\r\n", task.ToHttpHost())
	_, err = tlsConn.Write([]byte(httpReq))
	if err != nil {
		return "", nil, err
	}
	banner, err = common.ReadAll(tlsConn)
	if len(banner) > 0 && len(banner) > 5 && banner[0] == 'H' && banner[1] == 'T' && banner[2] == 'T' && banner[3] == 'P' && banner[4] == '/' {
		return "HTTPS", banner, nil
	} else {
		return "", banner, err
	}

}

func (s *HTTPS) DefaultPorts() []string {
	return []string{"443", "8443"}
}
func init() {
	RegistScanner(&HTTPS{})
}
