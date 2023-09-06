package services

import (
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"easm_punkmap/common"
	"fmt"
	"net"
)

type HTTPS struct {
}
type CertificateInfo struct {
	Subject       string   `json:"subject"`
	Issuer        string   `json:"issuer"`
	ValidFrom     string   `json:"valid_from"`
	ValidUntil    string   `json:"valid_until"`
	SerialNumber  string   `json:"serial_number"`
	DNSNames      []string `json:"dns_names"`
	KeyUsage      []string `json:"key_usage"`
	ExtKeyUsage   []string `json:"ext_key_usage"`
	SignatureAlgo string   `json:"signature_algorithm"`
	PublicKeyAlgo string   `json:"public_key_algorithm"`
}

func (s *HTTPS) Scan(conn net.Conn, task Task, result *Result) (service string, banner []byte, err error) {

	// convert conn to tls
	tlsConn := tls.Client(conn, &tls.Config{InsecureSkipVerify: true})
	httpReq := fmt.Sprintf("GET / HTTP/1.1\r\nUser-Agent: PunkMap (https://github.com/openeasm/punkmap)\r\nHost: %s\r\nConnection: close\r\nAccept: */*\r\n\r\n", task.ToHttpHost())
	_, err = tlsConn.Write([]byte(httpReq))
	if err != nil {
		return "", nil, err
	}
	banner, err = common.ReadAll(tlsConn)
	cert := tlsConn.ConnectionState().PeerCertificates[0]
	//convert cert to map[]
	var mapedCert = map[string]interface{}{}
	mapedCert["cert_subject"] = cert.Subject.String()
	mapedCert["cert_issuer"] = cert.Issuer.String()
	mapedCert["cert_valid_from"] = cert.NotBefore.String()
	mapedCert["cert_valid_until"] = cert.NotAfter.String()
	mapedCert["cert_serial_number"] = cert.SerialNumber.String()
	mapedCert["cert_dns_names"] = cert.DNSNames
	mapedCert["cert_key_usage"] = cert.KeyUsage
	mapedCert["cert_ext_key_usage"] = cert.ExtKeyUsage
	mapedCert["cert_signature_algorithm"] = cert.SignatureAlgorithm.String()
	mapedCert["cert_public_key_algorithm"] = cert.PublicKeyAlgorithm.String()
	mapedCert["cert_human_readable"] = x509certHumanReadable(cert)
	result.ServiceAddition = mapedCert
	if len(banner) > 0 && len(banner) > 5 && banner[0] == 'H' && banner[1] == 'T' && banner[2] == 'T' && banner[3] == 'P' && banner[4] == '/' {
		return "HTTPS", banner, nil
	} else {
		return "", banner, err
	}

}
func x509certHumanReadable(cert *x509.Certificate) string {
	certificateInfo := ""
	certificateInfo += fmt.Sprintf("Version: v%d\n", cert.Version)
	certificateInfo += fmt.Sprintf("Serial Number: %s\n", cert.SerialNumber)
	certificateInfo += fmt.Sprintf("Signature Algorithm: %s\n", cert.SignatureAlgorithm)

	certificateInfo += "Issuer:\n"
	certificateInfo += formatDN(cert.Issuer)

	certificateInfo += "Validity:\n"
	certificateInfo += fmt.Sprintf("Not Before: %s\n", cert.NotBefore.UTC().Format("2006-01-02 15:04 MST"))
	certificateInfo += fmt.Sprintf("Not After : %s\n", cert.NotAfter.UTC().Format("2006-01-02 15:04 MST"))

	certificateInfo += "Subject:\n"
	certificateInfo += formatDN(cert.Subject)

	certificateInfo += "\n"
	return certificateInfo
}
func formatDN(dn pkix.Name) string {
	certificateInfo := ""
	if len(dn.Country) > 0 {
		certificateInfo += fmt.Sprintf("Country: %s\n", dn.Country[0])
	}
	if len(dn.Province) > 0 {
		certificateInfo += fmt.Sprintf("Province: %s\n", dn.Province[0])
	}
	if len(dn.Organization) > 0 {
		certificateInfo += fmt.Sprintf("Organization: %s\n", dn.Organization[0])
	}
	return certificateInfo
}

func (s *HTTPS) DefaultPorts() []string {
	return []string{"443", "8443"}
}
func init() {
	RegistScanner(&HTTPS{})
}
