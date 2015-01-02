/**
Convenience Class to Marshall and Unmarshall received Description Headers
*/
package messages

import (
	"../ipaddr"
	"encoding/binary"
	"fmt"
)

const PING byte = 0x00
const PONG byte = 0x01
const PUSH byte = 0x40
const QUERY byte = 0x80
const QUERYHIT byte = 0x81

type DescHeader struct {
	DescID      [16]byte
	PayloadDesc byte
	TTL         byte
	Hops        byte
	PayloadLen  uint32
}

func (descHeader *DescHeader) EqualsID(otherID []byte) bool {
	if len(otherID) != len(descHeader.DescID) {
		return false
	}
	for i, v := range descHeader.DescID {
		if v != otherID[i] {
			return false
		}
	}
	return true
}

func (descHeader *DescHeader) Equals(otherHeader *DescHeader) bool {
	return descHeader.PayloadDesc == otherHeader.PayloadDesc && descHeader.EqualsID(otherHeader.DescID[:])
}

func parseHeaderBytes(rawHeader []byte, descHeader *DescHeader) error {
	if len(rawHeader) != 23 {
		return fmt.Errorf("input must be of length 23. Actual length == %d", len(rawHeader))
	}
	// Copy contents into member variables
	copy(descHeader.DescID[:], rawHeader[:16])
	descHeader.PayloadDesc = rawHeader[16]
	descHeader.TTL = rawHeader[17]
	descHeader.Hops = rawHeader[18]
	descHeader.PayloadLen = binary.LittleEndian.Uint32(rawHeader[19:])
}

func ParseHeaderBytes(rawHeader []byte) (*DescHeader, error) {
	descHeader := new(DescHeader)
	err := parseHeaderBytes(rawHeader, descHeader)
	return descHeader, err
}

func (descHeader *DescHeader) ParseHeaderBytes(rawHeader []byte) error {
	err := parseHeaderBytes(rawHeader, descHeader)
	return err
}

func (descHeader *DescHeader) ToBytes() []byte {
	var buffer [23]byte
	copy(buffer[:16], descHeader.DescID[:])
	buffer[16] = descHeader.PayloadDesc
	buffer[17] = descHeader.TTL
	buffer[18] = descHeader.Hops
	binary.LittleEndian.PutUint32(buffer[19:], descHeader.PayloadLen)
	return buffer[:]
}
