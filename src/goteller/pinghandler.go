package goteller

import (
	"../ipaddr"
	"../messages"
)

func (teller *GoTeller) onPing(descHeader messages.DescHeader, from ipaddr.IPAddr) {
	pong := messages.PongMsg{NumShared: teller.NumShared, NumKB: teller.NumKB}
	pong.Addr = teller.addr
	pongBuffer := pong.ToBytes()
	pongHeader := messages.DescHeader{
		DescID:      descHeader.DescID, // Very Important! Pong Must be same ID as Ping
		PayloadDesc: messages.PONG,
		TTL:         descHeader.Hops,
		PayloadLen:  uint32(len(pongBuffer)),
	}
	headerBuffer := pongHeader.ToBytes()
	msgBuffer := append(headerBuffer, pongBuffer...)
	teller.sendToNeighbor(msgBuffer, from)
	descHeader.TTL--
	descHeader.Hops++
	if descHeader.TTL > 0 {
		teller.pingMapMutex.Lock()
		teller.savedPings[descHeader.DescID] = from // Save in saved Pings
		teller.pingMapMutex.Unlock()
		teller.floodToNeighbors(descHeader.ToBytes(), from)
	}
}
