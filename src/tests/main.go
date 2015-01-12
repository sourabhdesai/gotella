package main

import (
	"./goteller"
	"fmt"
)

func main() {
	fmt.Println("Hi")
	teller := goteller.GoTeller{}
	fmt.Printf("%+v\n", teller)
}
