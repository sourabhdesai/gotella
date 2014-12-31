package ipaddr

import (
	"encoding/binary"
	"fmt"
)

type IPAddr struct {
	IP   [4]byte
	Port uint16
}

// Takes strings of format "a.b.c.d:p" and returns equivalent IPAddr struct
func parseString(addr string, ipAddr *IPAddr) error {
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

func ParseString(addr string) (*IPAddr, error) {
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
	copy(ipAddr.IP[:], rawAddr[2:])
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
	copy(buffer[2:], ipAddr.IP[:])
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

func Test1() {
	addr, err := ParseString("127.0.0.1:80")
	for i := 0; i < 10; i++ {
		if err != nil {
			fmt.Println(err)
		} else {
			if addr.IP[0] != byte(127) {
				fmt.Printf("Wrong value at idx 0. : %d... expected %o\n", addr.IP[0], uint8(127))
			}
			if addr.IP[1] != 0 {
				fmt.Printf("Wrong value at idx 1 : %d\n", addr.IP[1])
			}
			if addr.IP[2] != 0 {
				fmt.Printf("Wrong value at idx 2 : %d\n", addr.IP[2])
			}
			if addr.IP[3] != byte(1) {
				fmt.Printf("Wrong value at idx 3 : %d\n", addr.IP[3])
			}
			if addr.Port != uint16(80) {
				fmt.Printf("Wrong port value : %d\n", addr.Port)
			}
			fmt.Println(addr)
		}
		addr, err = ParseString(addr.String()) // Marshall then unmarshall
	}
}

func Test2() {
	addr, err := ParseString("127.0.0.1:80")
	if err != nil {
		fmt.Println(err)
		return
	}
	buff := addr.ToBytes()
	addr, err = ParseBytes(buff)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(addr)
}

/*
func main() {
	Test2()
}
*/
