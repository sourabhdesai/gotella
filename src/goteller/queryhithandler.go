package goteller

import (
	"../messages"
)

func (teller *GoTeller) onQueryHit(header messages.DescHeader, queryHit messages.QueryHitMsg) {
	teller.queryMapMutex.RLock()
	if querySrc, ok := teller.savedQueries[header.DescID]; ok {
		teller.queryMapMutex.RUnlock()
		if querySrc == teller.addr {
			// Query was from this node
			results := resultsFromHit(queryHit)
			chosenResults := teller.resultFunc(results, queryHit.Speed, string(queryHit.ServantID[:]))
			for _, result := range chosenResults {
				go teller.sendRequest(result.fileIndex, result.filename, result.addr)
			}
		} else if header.TTL > 0 {
			header.TTL--
			header.Hops++
			msgBuffer := append(header.ToBytes(), queryHit.ToBytes()...)
			teller.sendToNeighbor(msgBuffer, querySrc)
		}
	} else {
		teller.queryMapMutex.RUnlock()
	}
}
