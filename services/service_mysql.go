package services

import (
	"easm_punkmap/common"
	"net"
)

type MySQL struct {
}

func (s *MySQL) Scan(conn net.Conn, task Task, result *Result) (service string, banner []byte, err error) {
	// MySQL握手协议，见：https://dev.mysql.com/doc/dev/mysql-server/latest/page_protocol_basic_dt_integers.html#a_protocol_type_int1
	// Example Response Banner:
	// 00000000  4a 00 00 00 0a 35 2e 37  2e 34 32 00 27 00 00 00  |J....5.7.42.'...|
	// 00000010  5b 04 42 3f                                       |[.B?|
	// 包内容解释：
	// 4a: Packet length, indicating that the total length of the packet is 74 bytes.
	// 00 00 00: Filler bytes.
	// 0a: Sequence number.
	// 35 2e 37 2e 34 32 00: MySQL server version string, which seems to be "5.7.42."
	// 27: Connection ID.
	// 00 00 00: More filler bytes.
	// 5b 04 42 3f: Part of the authentication data.
	banner, err = common.ReadUntilNewLine(conn)
	if banner != nil {
		if banner[1] == 0x00 && banner[2] == 0x00 && banner[3] == 0x00 {
			return "mysql", banner, err
		} else {
			return "", banner, err
		}
	}
	return "", banner, err
}

func (s *MySQL) DefaultPorts() []string {
	return []string{"3306"}
}
func init() {
	RegistScanner(&MySQL{})
}
