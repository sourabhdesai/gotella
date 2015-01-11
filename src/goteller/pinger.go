package goteller

import (
	"fmt"
	"net"
	"time"
)

const DEFAULT_PING_TTL byte = 2

// Must be run on separate goroutine
func (teller *GoTeller) startPinger() {
	for teller.alive {
		teller.pingLoop(DEFAULT_PING_TTL)
	}
}

func (teller *GoTeller) pingLoop(ttl byte) {
	defer func() {
		if r := recover(); r != nil {
			if teller.debugFile != nil {
				fmt.Fprintln(teller.debugFile, r)
			}
		}
	}()

	// Sleep for a set interval period
	time.Sleep(teller.PingInterval)

	header := DescHeader{
		DescID:      teller.newID(),
		PayloadDesc: messages.PING,
		TTL:         ttl,
		Hops:        0,
		PayloadLen:  0x0000000000,
	}
	pingBuffer := header.ToBytes()
	teller.neighborsMutex.RLock()
	rlocked := true
	defer func() {
		if rlocked {
			teller.neighborsMutex.RUnlock()
		}
	}()
	for addr, i := range teller.Neighbors {
		sent := teller.sendToNeighbor
		if !sent { // !sent Indicates the neighbor is dead
			teller.neighborsMutex.RUnlock()
			rlocked = false
			teller.removeNeighbor(addr)
			teller.neighborsMutex.RLock()
			rlocked = true
		}
	}
}
