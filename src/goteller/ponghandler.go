package goteller

import (
	"../messages"
)

func (teller *GoTeller) onPong(header messages.DescHeader, pong messages.PongMsg) {
	teller.pingMapMutex.RLock()
	if pingSrc, ok := teller.savedPings[header.DescID]; ok {
		teller.pingMapMutex.RUnlock()
		if pingSrc == teller.addr {
			// Pong is for self
			if pong.Addr != teller.addr && !teller.isNeighbor(pong.Addr) { // Only add to neighbor list if its not already a neighbor and address isn't for self
				teller.addNeighbor(pong.Addr)
			}
		} else if header.TTL > 0 {
			header.TTL--
			header.Hops++
			msgBuffer := append(header.ToBytes(), pong.ToBytes()...)
			teller.sendToNeighbor(msgBuffer, pingSrc)
		} // If TTL is 0, do not forward.
		teller.pingMapMutex.Lock()
		defer teller.pingMapMutex.Unlock()
		delete(teller.savedPings, header.DescID) // Remove entry
	} else {
		teller.pingMapMutex.RUnlock()
	}
}
