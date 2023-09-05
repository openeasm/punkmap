package common

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

var titleRegex = regexp.MustCompile(`(?i)<title>(.*?)</title>`)
var keyRegex = regexp.MustCompile(`(?i)<meta.*?name=["']?keywords["']?.*?content=["']?(.*?)["']?/?>`)
var descRegex = regexp.MustCompile(`(?i)<meta.*?name=["']?description["']?.*?content=["']?(.*?)["']?/?>`)

func HTTPParser(banner []byte) map[string]string {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("http parser panic:", err)
			fmt.Println("------------banner start ------------")
			fmt.Println(string(banner))
			fmt.Println("------------banner end ------------")
		}
	}()
	result := make(map[string]string)
	if bytes.Contains(banner, []byte("\r\n\r\n")) {
		var bannerHeader = banner[:bytes.Index(banner, []byte("\r\n\r\n"))]
		headerLine := bytes.Split(bannerHeader, []byte("\r\n"))
		protocol := bytes.Split(headerLine[0], []byte(" "))
		result["http_version"] = string(protocol[0])
		result["http_status_code"] = string(protocol[1])
		result["http_headers"] = string(bannerHeader)
		for _, line := range headerLine[1:] {
			kv := bytes.Split(line, []byte(": "))
			if len(kv) == 2 {
				if strings.ToLower(string(kv[0])) == "server" {
					result["http_server"] = string(kv[1])
				}
			}
		}
	}
	if titleRegex.Match(banner) {
		result["http_title"] = string(titleRegex.FindSubmatch(banner)[1])
	}
	if keyRegex.Match(banner) {
		result["http_keywords"] = string(keyRegex.FindSubmatch(banner)[1])
	}
	if descRegex.Match(banner) {
		result["http_description"] = string(descRegex.FindSubmatch(banner)[1])
	}
	return result
}
