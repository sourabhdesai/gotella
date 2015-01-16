package goteller

import (
	"../messages"
)

func (teller *GoTeller) onQueryHit(header messages.DescHeader, queryHit messages.QueryHitMsg) {
	teller.myQueryMapMutex.RLock()
	if query, ok := teller.myQueries[header.DescID]; ok {
		// Query was from this node
		teller.myQueryMapMutex.RUnlock()
		results := resultsFromHit(queryHit)
		chosenResults := query.onHit(results, queryHit.Speed, string(queryHit.ServantID[:]))
		for _, result := range chosenResults {
			go teller.sendRequest(result.fileIndex, result.filename, result.addr, query.onResponse)
		}
	} else {
		// Is not your own query... must forward to appropriate neighbor
		teller.myQueryMapMutex.RUnlock()
		teller.queryMapMutex.RLock()
		if querySrc, ok := teller.savedQueries[header.DescID]; ok {
			teller.queryMapMutex.RUnlock()
			if header.TTL > 0 { // Only forward if TTL > 0
				header.TTL--
				header.Hops++
				msgBuffer := append(header.ToBytes(), queryHit.ToBytes()...)
				teller.sendToNeighbor(msgBuffer, querySrc)
			}
		} else {
			teller.queryMapMutex.RUnlock()
		}
	}
}
