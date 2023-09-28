package common

import (
	"bytes"
	"io"
	"net"
)

func ReadAll(conn net.Conn) (data []byte, err error) {
	// read data from conn
	buffer := make([]byte, 1024)

	for {
		n, err := conn.Read(buffer)
		data = append(data, buffer[:n]...)
		if err != nil {
			//if timeout,return no error
			if err.Error() == "EOF" {
				err = nil
			}
			return data, err
		}
		if n < 1024 {
			return data, nil
		}
	}
}

func ReadUntilNewLine(conn net.Conn) (data []byte, err error) {
	buffer := make([]byte, 1024)
	readCnt := 0
	for {
		n, err := conn.Read(buffer)
		data = append(data, buffer[:n]...)
		readCnt += n
		if err == io.EOF {
			return data, nil
		} else if err != nil {
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
	for {
		n, err := conn.Read(buffer)
		data = append(data, buffer[:n]...)
		readCnt += n
		if err != nil {
			return data, err
		}
		if readCnt == maxBytes {
			return data, nil
		}
	}
}
