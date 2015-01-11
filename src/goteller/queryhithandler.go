package gotella

import (
	"../ipaddr"
	"../messages"
	"fmt"
)

func (teller *GoTeller) onQueryHit(header DescHeader, queryHit QueryHitMsg) {
	teller.queryMapMutex.RLock()
	if addr, ok := teller.savedQueries[header.DescID]; ok {
		teller.queryMapMutex.RUnlock()
		if queryHit.Addr == teller.addr {
			// Query was from this node
			results := resultsFromHit(queryHit)
			chosenResults := teller.resultFunc(results, queryHit.Speed, string(queryHit.ServantID))
			for _, result := range chosenResults {
				go teller.sendRequest(result.fileIndex, result.filename, result.addr)
			}
		} else if header.TTL > 0 {
			header.TTL--
			header.Hops++
			msgBuffer := append(header.ToBytes(), queryHit.ToBytes())
			teller.queryMapMutex.Lock()
			delete(teller.savedQueries, header.DescID) // remove entry
			teller.queryMapMutex.Unlock()
			teller.sendToNeighbor(msgBuffer, addr)
		}
	} else {
		teller.queryMapMutex.RUnlock()
	}
}
