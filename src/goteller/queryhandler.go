package goteller

import (
	"../ipaddr"
	"../messages"
	"fmt"
)

func (teller *GoTeller) onQuery(header *DescHeader, query *QueryMsg, from IPAddr) {
	if teller.NetworkSpeed >= query.MinSpeed {
		// This node meets speed requirements for query
		hitResults := teller.queryFunc(query.SearchQuery)
		if len(hitResults) > 0 {
			// Found results for given query
			queryHit := QueryHitMsg{
				NumHits:   len(hitResults),
				Addr:      teller.addr,
				Speed:     teller.NetworkSpeed,
				ResultSet: hitResults,
				ServantID: teller.ID,
			}
			queryHitBuffer := queryHit.ToBytes()
			queryHitHeader := DescHeader{
				DescID:      header.DescID,
				PayloadDesc: messages.QueryHit,
				TTL:         header.Hops,
				Hops:        0,
				PayloadLen:  len(queryHitBuffer),
			}
			headerBuffer := queryHitHeader.ToBytes()
			msgBuffer := append(headerBuffer, queryHitBuffer)
			teller.sendToNeighbor(msgBuffer, from)
		}
	}
	// Forward query to neighbors if TTL > 0
	if header.TTL > 0 {
		header.TTL--
		header.Hops++
		msgBuffer := append(header.ToBytes(), query.ToBytes())
		teller.floodToNeighbors(msgBuffer, from)
	}
}

func (teller *GoTeller) sendQuery(ttl byte, searchQuery string, minSpeed uint32, from IPAddr) {
	query := QueryMsg{
		MinSpeed:    minSpeed,
		SearchQuery: searchQuery,
	}
	queryBuffer := query.ToBytes()
	header := DescHeader{
		DescID:      teller.newID(),
		PayloadDesc: messages.QUERY,
		TTL:         ttl,
		Hops:        0,
		PayloadLen:  len(queryBuffer),
	}
	headerBuffer := header.ToBytes()
	msgBuffer := append(headerBuffer, queryBuffer)
	teller.floodToNeighbors(msgBuffer, from)
}
