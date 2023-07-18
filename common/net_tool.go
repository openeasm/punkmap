package common

import "net"

func ReadAll(conn net.Conn) (data []byte, err error) {
	// read data from conn
	buffer := make([]byte, 1024)
	var wait_ttl = 1
	for {
		n, err := conn.Read(buffer)
		data = append(data, buffer[:n]...)
		if err != nil && wait_ttl == 0 {
			return data, err
		}
		if n < 1024 && wait_ttl == 0 {
			return data, nil
		}
		wait_ttl--
	}
}
