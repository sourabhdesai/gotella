package main

import (
	"fmt"
	"../ipaddr"
)

func test1() {
	addr, err := ipaddr.ParseAddrString("127.0.0.1:8000")
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
			if addr.Port != uint16(8000) {
				fmt.Printf("Wrong port value : %d\n", addr.Port)
			}
			fmt.Println(addr)
		}
		addr, err = ipaddr.ParseAddrString(addr.String()) // Marshall then unmarshall
	}
}

func test2() {
	addr, err := ipaddr.ParseAddrString("127.0.0.1:8000")
	if err != nil {
		fmt.Println(err)
		return
	}
	buff := addr.ToBytes()
	addr, err = ipaddr.ParseBytes(buff)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(addr)
}

func main() {
	test1()
}