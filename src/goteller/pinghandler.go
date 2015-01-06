package gotella

import (
	"../messages"
	"fmt"
	"net"
)

func (teller *GoTeller) onPing(descHeader *DescHeader, from IPAddr) {
	pong := PongMsg{NumShared: teller.NumShared, NumKB: teller.NumKB}
	pong.Addr.Port = teller.Port
	pongBuffer := pong.ToBytes()
	pongHeader := DescHeader{
		DescID:      descHeader.DescID,
		PayloadDesc: messages.PONG,
		TTL:         descHeader.Hops,
		PayloadLen:  len(pongBuffer),
	}
	headerBuffer := pongHeader.ToBytes()
	msgBuffer := append(headerBuffer, pongBuffer)
	teller.sendToNeighbor(msgBuffer, from)
	descHeader.TTL--
	descHeader.Hops++
	if descHeader.TTL > 0 {
		teller.floodToNeighbors(descHeader.ToBytes(), from)
	}
}
