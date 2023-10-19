package services

import (
	"bytes"
	"easm_punkmap/common"
	"encoding/binary"
	"fmt"
	"net"
)

type MySQL struct {
}

func (m MySQL) Scan(conn net.Conn, task Task, result *Result) PortInfo {
	// MySQL握手协议，见：https://dev.mysql.com/doc/dev/mysql-server/latest/page_protocol_basic_dt_integers.html#a_protocol_type_int1
	// Example Response Banner:
	// 00000000  4a 00 00 00 0a 35 2e 37  2e 34 32 00 27 00 00 00  |J....5.7.42.'...|
	// 00000010  5b 04 42 3f                                       |[.B?|
	// 包内容解释：
	// 4a : Packet length, indicating that the total length of the packet is 74 bytes.
	// 00 00 00: Filler bytes.
	// 0a: Sequence number.
	// 35 2e 37 2e 34 32 00: MySQL server version string, which seems to be "5.7.42."
	// 27: Connection ID.
	// 00 00 00: More filler bytes.
	// 5b 04 42 3f: Part of the authentication data.
	lengthBuffer := make([]byte, 4)
	portInfo := PortInfo{}
	_, err := conn.Read(lengthBuffer)
	if err != nil {
		portInfo.err = err
		return portInfo
	}
	// 读取包长度，Little-endian
	packetLength := binary.LittleEndian.Uint32(lengthBuffer)

	// 读取包内容
	banner, err := common.ReadUntilNBytes(conn, int(packetLength))
	if err != nil {
		portInfo.banner = banner
		portInfo.err = err
		return portInfo
	}
	rawBanner := make([]byte, len(banner)+len(lengthBuffer))
	rawBanner = append(lengthBuffer, banner...)
	if banner != nil && len(banner) > 0 {
		banner = banner[1:]
		versionEnd := bytes.IndexByte(banner, 0x00)
		if versionEnd > 0 {
			mysqlVersion := string(banner[:versionEnd])
			portInfo.version = mysqlVersion
			portInfo.banner = rawBanner
			portInfo.service = "mysql"
			return portInfo
		} else {
			portInfo.banner = banner
			portInfo.err = fmt.Errorf("invalid mysql banner")
			return portInfo
		}
	}
	return portInfo
}

func (m MySQL) DefaultPorts() []string {
	//TODO implement me
	return []string{"3306"}
}

func init() {
	RegistScanner(&MySQL{})
}
