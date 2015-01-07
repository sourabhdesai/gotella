package goteller

import (
	"../ipaddr"
	"../messages"
	"fmt"
)

func (teller *GoTeller) onPong(header *DescHeader, pong *PongMsg) {
	teller.pingMapMutex.RLock()
	if from, ok := teller.savedPings[header.DescID]; ok {
		teller.pingMapMutex.RUnock()
		if from == teller.addr {
			// Pong is for self
			if !teller.isNeighbor(pong.Addr) {
				teller.addNeighbor(pong.Addr)
			}
		} else if header.TTL > 0 {
			header.TTL--
			header.Hops++
			msgBuffer := append(header.ToBytes(), pong.ToBytes())
			teller.sendToNeighbor(msgBuffer, from)
		} // If TTL is 0, do not forward.
		teller.pingMapMutex.Lock()
		defer teller.pingMapMutex.Unlock()
		delete(teller.savePings, header.DescID) // Remove entry
	} else {
		teller.pingMapMutex.RUnlock()
	}
}
