package network

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"time"
)

var PATTERN = regexp.MustCompile(`((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?\.){3})(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)`)

func IsFoundHost(host string, port uint16) bool {
	target := fmt.Sprintf("%s:%d", host, port)

	_, err := net.DialTimeout("tcp", target, 1*time.Second)
	if err != nil {
		fmt.Printf("%s %v\n", target, err)
		return false
	}
	return true
}

func FindNeighbors(myHost string, port uint16, startIP, endIP uint8, startPort, endPort uint16) ([]string, error) {
	address := fmt.Sprintf("%s:%d", myHost, port)
	m := PATTERN.FindStringSubmatch(address)
	if m == nil {
		return nil, fmt.Errorf("failed to substring regex from address: %s", address)
	}
	prefixHost := m[1]
	lastIP, err := strconv.Atoi(m[len(m)-1])
	if err != nil {
		return nil, err
	}
	neighbors := make([]string, 0)

	for port := startPort; port <= endPort; port += 1 {
		for ip := startIP; ip <= endIP; ip += 1 {
			guessHost := fmt.Sprintf("%s%d", prefixHost, lastIP+int(ip))
			guessTarget := fmt.Sprintf("%s:%d", guessHost, port)
			if guessTarget != address && IsFoundHost(guessHost, port) {
				neighbors = append(neighbors, guessTarget)
			}
		}
	}

	return neighbors, nil
}

func GetHost() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "127.0.0.1"
	}
	address, err := net.LookupHost(hostname)
	if err != nil {
		return "127.0.0.1"
	}

	if len(address) == 1 {
		return address[0]
	}

	return address[len(address)-1]
}
