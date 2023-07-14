package common

import "net"

func ReadAll(conn net.Conn) (data []byte, err error) {
	// read data from conn
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			return data, err
		}
		data = append(data, buffer[:n]...)
		if n < 1024 {
			return data, nil
		}
	}
}
