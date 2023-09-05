package services

import (
	"easm_punkmap/common"
	"fmt"
	"net"
)

type HTTP struct {
}

func (s *HTTP) Scan(conn net.Conn, task Task) (service string, banner []byte, err error) {
	// no need send data, read data directly
	// read data from conn
	httpReq := fmt.Sprintf("GET / HTTP/1.1\r\nUser-Agent: PunkMap (https://github.com/openeasm/punkmap)\r\nHost: %s\r\nConnection: close\r\nAccept: */*\r\n\r\n", task.ToHttpHost())
	_, err = conn.Write([]byte(httpReq))
	if err != nil {
		return "", nil, err
	}
	banner, err = common.ReadAll(conn)
	if len(banner) > 0 && len(banner) > 5 && banner[0] == 'H' && banner[1] == 'T' && banner[2] == 'T' && banner[3] == 'P' && banner[4] == '/' {
		return "HTTP", banner, nil
	} else {
		return "", banner, err
	}
}

func (s *HTTP) DefaultPorts() []string {
	return []string{"80", "8080", "8081", "8082", "8083", "8084", "8085", "8086", "8087", "8088", "8089", "8090", "8091"}
}
func init() {
	RegistScanner(&HTTP{})
}
