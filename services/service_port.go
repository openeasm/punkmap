package services

import "net"

type PortInfo struct {
	protocol string
	service  string
	banner   []byte
	version  string
	err      error
}

type ServiceScanner interface {
	Scan(conn net.Conn, task Task, result *Result) PortInfo
	DefaultPorts() []string
}

var PortScannersMapping = map[string][]ServiceScanner{}

func RegistScanner(scanner ServiceScanner) {
	for _, port := range scanner.DefaultPorts() {
		if PortScannersMapping[port] == nil {
			PortScannersMapping[port] = []ServiceScanner{}
		}
		PortScannersMapping[port] = append(PortScannersMapping[port], scanner)
	}
}
