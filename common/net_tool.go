package common

import (
	"bytes"
	"net"
)

func ReadAll(conn net.Conn) (data []byte, err error) {
	// read data from conn
	buffer := make([]byte, 1024)

	for waitTtl := 1; waitTtl > 0; waitTtl-- {
		n, err := conn.Read(buffer)
		data = append(data, buffer[:n]...)
		if err != nil {
			return data, err
		}
		if n < 1024 {
			return data, nil
		}
	}
	return data, err
}
func ReadUntilNewLine(conn net.Conn) (data []byte, err error) {
	buffer := make([]byte, 1)
	readCnt := 0
	for waitTtl := 1; waitTtl > 0; waitTtl-- {
		n, err := conn.Read(buffer)
		data = append(data, buffer[:n]...)
		readCnt += n
		if err != nil && waitTtl == 0 {
			return data, err
		}
		if bytes.Contains(buffer[:n], []byte("\n")) {
			return data, nil
		}
	}
}
func ReadUntilNBytes(conn net.Conn, maxBytes int) (data []byte, err error) {
	buffer := make([]byte, maxBytes)
	readCnt := 0
	for waitTtl := 1; waitTtl > 0; waitTtl-- {
		n, err := conn.Read(buffer)
		data = append(data, buffer[:n]...)
		readCnt += n
		if err != nil && waitTtl == 0 {
			return data, err
		}
		if readCnt >= maxBytes {
			return data, nil
		}
	}
}
