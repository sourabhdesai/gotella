package gotella

import (
	"../messages"
	"fmt"
	"net"
)

func (teller *GoTeller) onPing(descHeader DescHeader, from IPAddr) {
	pong := PongMsg{NumShared: teller.NumShared, NumKB: teller.NumKB}
	pong.Addr = teller.addr
	pongBuffer := pong.ToBytes()
	pongHeader := DescHeader{
		DescID:      descHeader.DescID, // Very Important! Pong Must be same ID as Ping
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
		teller.pingMapMutex.Lock()
		teller.savedPings[descHeader.DescID] = from // Save in saved Pings
		teller.pingMapMutex.Unlock()
	}
}

func (teller *GoTeller) sendPings(ttl byte) {
	header := DescHeader{
		DescID:      teller.newID(),
		PayloadDesc: messages.PING,
		TTL:         ttl,
		Hops:        0,
		PayloadLen:  0x0000000000,
	}
	msgBuffer := header.ToBytes()
	teller.floodToNeighbors(msgBuffer, teller.Addr)
}
