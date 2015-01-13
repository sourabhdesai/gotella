package main

import (
	"fmt"
	"../messages"
)

func queryHitsEqual(a, b messages.QueryHitMsg) bool {
	if len(a.ResultSet) != len(b.ResultSet) {
		return false
	}
	for i,_ := range a.ResultSet {
		if a.ResultSet[i] != b.ResultSet[i] {
			return false
		}
	}

	return (a.NumHits == b.NumHits) &&
			(a.Speed == b.Speed) &&
			(a.ServantID == b.ServantID) &&
			(a.Addr == b.Addr)

}

func TestByteSerialization() {
	queryHit := messages.QueryHitMsg {
		NumHits: 1,
		Speed: 20,
	}
	copy(queryHit.ServantID[:], []byte("sourabhdesai1993"))
	queryHit.Addr.SetToLocalIP()
	queryHit.ResultSet = []messages.HitResult{
		messages.HitResult{
			FileIndex: 69,
			FileSize: 99,
			Filename: "hi.txt",
		},
	}
	fmt.Printf("Original:\n %+v\n", queryHit)

	queryHitBuffer := queryHit.ToBytes()

	queryHitCopy, err := messages.ParseQueryHitBytes(queryHitBuffer)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Copy:\n %+v\n", queryHitCopy)
	fmt.Printf("%t\n", queryHitsEqual(queryHit, *queryHitCopy) )
} 

func main() {
	TestByteSerialization()
}