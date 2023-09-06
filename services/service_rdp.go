package services

import (
	"easm_punkmap/common"
	"net"
)

type RDP struct {
}

func (s *RDP) Scan(conn net.Conn, task Task, result *Result) (service string, banner []byte, err error) {
	// read data from conn
	/* 0x03,       # version
	   0x00,       # reserved
	   0x00, 0x13, # length (19)
	   0x0e,       # length (14)
	   0xe0,       # PDU type
	   0x00, 0x00, # dest ref
	   0x00, 0x00, # source ref
	   0x00,       # class
	   0x01,       # RDP negotiation request: type (1 byte): An 8-bit, unsigned integer that indicates the packet type. This field MUST be set to 0x01 (TYPE_RDP_NEG_REQ).
	   0x00,       # flags
	   0x08, 0x00, # length (8) # length (2 bytes): A 16-bit, unsigned integer that specifies the packet size. This field MUST be set to 0x0008 (8 bytes).
	   0x03, 0x00, 0x00, 0x00 # requested protocols. TLS security supported: TRUE, CredSSP supported: TRUE, EUARPDUS: FALSE
	*/
	rdpHandshake := []byte{0x03, 0x00, 0x00, 0x13, 0x0e, 0xe0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x08, 0x00, 0x03, 0x00, 0x00, 0x00}
	_, err = conn.Write(rdpHandshake)
	if err != nil {
		return "", nil, err
	}
	banner, err = common.ReadUntilNBytes(conn, 19)
	if err == nil {
		if len(banner) == 19 && banner[0] == 0x03 && banner[1] == 0x00 && banner[2] == 0x00 {
			return "RDP", banner, nil
		}
	}
	return "", banner, err
}

func (s *RDP) DefaultPorts() []string {
	return []string{"3389"}
}
func init() {
	RegistScanner(&RDP{})
}
