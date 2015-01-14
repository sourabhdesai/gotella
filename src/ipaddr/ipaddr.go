package ipaddr

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

type IPAddr struct {
	IP   [4]byte
	Port uint16
}

// Takes strings of format "a.b.c.d:p" and returns equivalent IPAddr struct
func parseString(addr string, ipAddr *IPAddr) error {
	if strings.HasPrefix(addr, "localhost:") { // Can parse address string in format localhost:port
		n, err := fmt.Sscanf(addr, "localhost:%d", &ipAddr.Port)
		if err != nil {
			return err
		}
		if n != 1 {
			return fmt.Errorf("Input string \"%s\" wasn't correct format", addr)
		}
		err = ipAddr.SetToLocalIP()
		return err // could be nil
	} else {
		n, err := fmt.Sscanf(addr, "%d.%d.%d.%d:%d",
			&ipAddr.IP[0],
			&ipAddr.IP[1],
			&ipAddr.IP[2],
			&ipAddr.IP[3],
			&ipAddr.Port)
		if err != nil {
			return err
		}
		if n != 5 {
			return fmt.Errorf("Input string \"%s\" wasn't correct format", addr)
		}
		return nil
	}
}

func ParseAddrString(addr string) (*IPAddr, error) {
	ipAddr := new(IPAddr)
	err := parseString(addr, ipAddr)
	return ipAddr, err
}

func (ipAddr *IPAddr) ParseString(addr string) error {
	err := parseString(addr, ipAddr)
	return err
}

// Takes buffer of 6 bytes. First two are for port, remaining are for IP ... like in Gnutella pong
func parseBytes(rawAddr []byte, ipAddr *IPAddr) error {
	if len(rawAddr) != 6 {
		return fmt.Errorf("Expected input buffer of length 6. Actualy length was %d", len(rawAddr))
	}
	ipAddr.Port = binary.LittleEndian.Uint16(rawAddr[:2])
	ipBEBuffer := bytes.NewReader(rawAddr[2:])
	binary.Read(ipBEBuffer, binary.BigEndian, ipAddr.IP[:])
	return nil
}

func ParseBytes(rawAddr []byte) (*IPAddr, error) {
	ipAddr := new(IPAddr)
	err := parseBytes(rawAddr, ipAddr)
	return ipAddr, err
}

func (ipAddr *IPAddr) ParseBytes(rawAddr []byte) error {
	err := parseBytes(rawAddr, ipAddr)
	return err
}

func (ipAddr *IPAddr) ToBytes() []byte {
	var buffer [6]byte
	binary.LittleEndian.PutUint16(buffer[:2], ipAddr.Port)
	buffWriter := new(bytes.Buffer)
	binary.Write(buffWriter, binary.BigEndian, ipAddr.IP[:])
	copy(buffer[2:], buffWriter.Bytes())
	return buffer[:]
}

func (ipAddr IPAddr) String() string {
	return fmt.Sprintf("%d.%d.%d.%d:%d",
		ipAddr.IP[0],
		ipAddr.IP[1],
		ipAddr.IP[2],
		ipAddr.IP[3],
		ipAddr.Port)
}

func LocalIPAtPort(port uint16) (*IPAddr, error) {
	ipString, err0 := getLocalIP()
	if err0 != nil {
		return nil, err0
	}
	addrString := fmt.Sprintf("%s:%d", ipString, port)
	addr, err1 := ParseAddrString(addrString)
	if err1 != nil {
		return nil, err1
	}
	return addr, nil
}

func (ipAddr *IPAddr) SetToLocalIP() error {
	ipString, err0 := getLocalIP()
	if err0 != nil {
		return err0
	}
	addrString := fmt.Sprintf("%s:%d", ipString, ipAddr.Port)
	err1 := ipAddr.ParseString(addrString)
	return err1 // Could be nil
}
