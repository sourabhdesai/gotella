package goteller

import (
	"../ipaddr"
	"../messages"
	"fmt"
)

func (teller *GoTeller) onQuery(header messages.DescHeader, query messages.QueryMsg, from ipaddr.IPAddr) {
	if from == teller.addr {
		return
	}
	// First check if we've seen this query before
	teller.queryMapMutex.RLock()
	if _, seen := teller.savedQueries[header.DescID]; seen {
		teller.queryMapMutex.RUnlock()
		return // No need to do anything further. Just drop it
	} else { // Haven't seen it.
		teller.queryMapMutex.RUnlock()
	}

	if teller.NetworkSpeed >= uint32(query.MinSpeed) {
		// This node meets speed requirements for query
		hitResults := teller.queryFunc(query.SearchQuery)
		if len(hitResults) > 0 {
			// Found results for given query
			var id [16]byte
			copy(id[:], []byte(teller.servantID))
			queryHit := messages.QueryHitMsg{
				NumHits:   byte(len(hitResults)),
				Addr:      teller.addr,
				Speed:     teller.NetworkSpeed,
				ResultSet: []messages.HitResult(hitResults),
				ServantID: id,
			}
			queryHitBuffer := queryHit.ToBytes()
			queryHitHeader := messages.DescHeader{
				DescID:      header.DescID,
				PayloadDesc: messages.QUERYHIT,
				TTL:         header.Hops,
				Hops:        0,
				PayloadLen:  uint32(len(queryHitBuffer)),
			}
			headerBuffer := queryHitHeader.ToBytes()
			msgBuffer := append(headerBuffer, queryHitBuffer...)
			sent := teller.sendToNeighbor(msgBuffer, from) // Send hit to neighbor
			if !sent {
				if teller.debugFile != nil {
					fmt.Fprintln(teller.debugFile, "Couldn't send QueryHitMsg to neighbor at "+from.String())
				}
			}
		}
	}
	// Forward query to neighbors if TTL > 0
	if header.TTL > 0 {
		header.TTL--
		header.Hops++
		msgBuffer := append(header.ToBytes(), query.ToBytes()...)
		teller.queryMapMutex.Lock()
		teller.savedQueries[header.DescID] = from // Save to savedQueries map
		teller.queryMapMutex.Unlock()
		teller.floodToNeighbors(msgBuffer, from)
	} else {
		teller.queryMapMutex.RUnlock()
	}
}

func (teller *GoTeller) sendQuery(searchQuery string, ttl byte, minSpeed uint16, from ipaddr.IPAddr) [16]byte {
	query := messages.QueryMsg{
		MinSpeed:    minSpeed,
		SearchQuery: searchQuery,
	}
	queryBuffer := query.ToBytes()
	header := messages.DescHeader{
		DescID:      teller.newID(),
		PayloadDesc: messages.QUERY,
		TTL:         ttl,
		Hops:        0,
		PayloadLen:  uint32(len(queryBuffer)),
	}
	headerBuffer := header.ToBytes()
	msgBuffer := append(headerBuffer, queryBuffer...)
	if from != teller.addr {
		teller.queryMapMutex.Lock()
		teller.savedQueries[header.DescID] = from // Save to savedQueries map
		teller.queryMapMutex.Unlock()
	}
	teller.floodToNeighbors(msgBuffer, from)
	return header.DescID
}
