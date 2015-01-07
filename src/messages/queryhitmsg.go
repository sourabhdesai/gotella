package messages

import (
	"../ipaddr"
	"encoding/binary"
	"fmt"
)

type HitResult struct {
	FileIndex uint32
	FileSize  uint32
	Filename  string
}

type QueryHitMsg struct {
	NumHits   byte
	Addr      IPAddr
	Speed     uint32
	ResultSet []HitResult
	ServantID [16]byte
}

func findDoubleNullByte(buffer []byte) int {
	for i := len(buffer) - 2; i >= 0; i-- {
		if buffer[i] == 0x00 && buffer[i+1] == 0x00 {
			return i
		}
	}
	return -1
}

func parseHitResultBytes(buffer []byte, hit *HitResult) error {
	if len(buffer) <= 10 {
		return fmt.Errorf("Expected buffer of length > 10. Got input buffer of length %d", len(buffer))
	}
	hit.FileIndex = binary.LittleEndian.Uint32(buffer[:4])
	hit.FileSize = binary.LittleEndian.Uint32(buffer[4:8])
	filenameBuffer := buffer[8:]
	nullIdx := findDoubleNullByte(filenameBuffer)
	if nullIdx == -1 {
		return fmt.Errorf("Couldn't find double null character in input buffer")
	}
	var err error = nil
	hit.Filename, err = readStringLE(filenameBuffer[:nullIdx])
	return err // err might be nil
}

func ParseHitResultBytes(buffer []byte) (*HitResult, error) {
	hit := new(HitResult)
	err := parseHitResultBytes(buffer, hit)
	return err
}

func (hit *HitResult) ParseBytes(buffer []byte) error {
	err := parseHitResultBytes(buffer, hit)
	return err
}

func (hit *HitResult) ByteLength() int {
	return 10 + len(hit.Filename) // 4 for Fileindex, 4 for Filesize, 2 for double null termination, len(hit.Filename)
}

func (hit *HitResult) ToBytes() []byte {
	bufferLen := hit.ByteLength()
	buffer := make([]byte, bufferLen)
	binary.LittleEndian.PutUint32(buffer[:4], hit.FileIndex)
	binary.LittleEndian.PutUint32(buffer[4:8], hit.FileSize)
	err := writeStringLE(buffer[:bufferLen-2], hit.Filename)
	buffer[bufferLen-2] = 0x00
	buffer[bufferLen-1] = 0x00
	return buffer
}

func parseQueryHitBytes(buffer []byte, queryHit *QueryHitMsg) error {
	if len(buffer) <= 21 {
		return fmt.Errorf("Expected buffer of length > 21. Input buffer was length %d", len(buffer))
	}
	queryHit.NumHits = buffer[0]
	err0 := queryHit.Addr.ParseBytes(buffer[1:7])
	if err0 != nil {
		return err0
	}
	queryHit.Speed = binary.LittleEndian.Uint32(buffer[7:11])
	queryHit.ResultSet = make([]HitResult, 0, queryHit.NumHits) // slice with 0 len, NumHits cap
	// Parse the Results Set
	var hitIdx int = 11
	for i := 0; i < queryHit.NumHits; i++ {
		if hitIdx > len(buffer)-16 { // Last 16 bytes are for servent identifier
			return fmt.Errorf("Number of Hits indicated doesn't match number of hits given")
		}
		hit, err1 := ParseHitResultBytes(buffer[hitIdx:])
		if err1 != nil {
			return err1
		}
		append(queryHit.ResultSet, *hit)
		hitIdx += hit.ByteLength()
	}
	var err1 error = nil
	queryHit.ServantID, err1 = readStringLE(buffer[hitIdx:])
	return er1r // err1 might be nil
}

func ParseQueryHitBytes(buffer []byte) (*QueryHitMsg, error) {
	queryHit := new(QueryHitMsg)
	err := parseQueryHitBytes(buffer, queryHit)
	return queryHit, err
}

func (queryHit *QueryHitMsg) ParseBytes(buffer []byte) error {
	err := parseQueryHitBytes(buffer, queryHit)
	return err
}

func (queryHit *QueryHitMsg) ByteLength() int {
	var hitResultsLength int = 0
	for _, hit := range queryHit.ResultSet {
		hitResultsLength += hit.ByteLength()
	}
	return 27 + hitResultsLength
}

func (queryHit *QueryHitMsg) ToBytes() []byte {
	bufferLen := queryHit.ByteLength()
	buffer := make([]byte, bufferLen)
	buffer[0] = queryHit.NumHits
	addrBytes := queryHit.Addr.ToBytes()
	copy(buffer[1:7], addrBytes)
	binary.LittleEndian.PutUint32(buffer[7:11], queryHit.Speed)
	hitIdx := 11
	for i := 0; i < len(queryHit.ResultSet); i++ {
		hitBytes := queryHit.ResultSet[i].ToBytes()
		copy(buffer[hitIdx:], hitBytes)
		hitIdx += len(hitBytes)
	}
	err := writeStringLE(buffer[hitIdx:], queryHit.ServantID)
	return buffer
}
