package services

import "net"

type ServiceScanner interface {
	Scan(conn net.Conn) (service string, banner []byte, err error)
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
