package ipaddr

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
)

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		return nil, err
	}

	for _, address := range addrs {

		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return nil, fmt.Errorf("Couldn't find a valid local IP address")
}
