package messages

import (
	"../ipaddr"
	"encoding/binary"
	"fmt"
)

type PongMsg struct {
	Addr      ipaddr.IPAddr
	NumShared uint32
	NumKB     uint32
}

func parsePongBytes(buffer []byte, pong *PongMsg) error {
	if len(buffer) != 14 {
		return fmt.Errorf("Expected buffer of length 14. Received buffer of length %d", len(buffer))
	}
	err0 := pong.Addr.ParseBytes(buffer[:6])
	if err0 != nil {
		return err0
	}
	pong.NumShared = binary.LittleEndian.Uint32(buffer[6:10])
	pong.NumKB = binary.LittleEndian.Uint32(buffer[10:])
	return nil
}

func ParsePongBytes(buffer []byte) (*PongMsg, error) {
	pong := new(PongMsg)
	err := parsePongBytes(buffer, pong)
	return pong, err
}

func (pong *PongMsg) ParseBytes(buffer []byte) error {
	err := parsePongBytes(buffer, pong)
	return err
}

func (pong *PongMsg) ToBytes() []byte {
	var buffer [14]byte
	addrBytes := pong.Addr.ToBytes() // Always of length 6
	copy(buffer[:6], addrBytes)
	binary.LittleEndian.PutUint32(buffer[6:10], pong.NumShared)
	binary.LittleEndian.PutUint32(buffer[10:], pong.NumKB)
	return buffer[:]
}
