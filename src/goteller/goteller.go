package goteller

import (
	"../ipaddr"
	"fmt"
	"net"
)

type GoTeller struct {
	addr      IPAddr
	Neighbors []IPAddr
	NumShared uint32
	NumKB     uint32
	Port      uint16
	hashCount uint32
	ping
}

func (teller *GoTeller) floodToNeighbors(msg []byte, from IPAddr) {
	//TODO: Implement this. Send msg to all neighbors except for from
}

func (teller *GoTeller) sendToNeighbor(msg []byte, from IPAddr) {
	//TODO: Implement this. Send message to neighbor
}

func (teller *GoTeller) newHash() [16]byte {
	var hash [16]byte
	addrBuffer := teller.addr.ToBytes()
	copy(hash[:6], addrBuffer)
}
