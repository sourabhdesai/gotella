package messages

import (
	"../ipaddr"
	"encoding/binary"
	"fmt"
)

type PushMsg struct {
	ServantID string
	FileIndex uint32
	Addr      ipaddr.IPAddr
}

func parsePushBytes(buffer []byte, push *PushMsg) error {
	if len(buffer) != 26 {
		return fmt.Errorf("Expected buffer of length 26. Got buffer of length %d", len(buffer))
	}
	push.ServantID = string(buffer[:16])
	push.FileIndex = binary.LittleEndian.Uint32(buffer[16:20])
	addrBuffer := make([]byte, 6)
	copy(addrBuffer[:2], buffer[24:])
	copy(addrBuffer[2:], buffer[20:24])
	err := push.Addr.ParseBytes(addrBuffer)
	return err // could be nil
}

func ParsePushBytes(buffer []byte) (*PushMsg, error) {
	push := new(PushMsg)
	err := parsePushBytes(buffer, push)
	return push, err
}

func (push *PushMsg) ParseBytes(buffer []byte) error {
	err := parsePushBytes(buffer, push)
	return err
}

func (push *PushMsg) ToBytes() []byte {
	buffer := make([]byte, 26)
	copy(buffer[:16], push.ServantID)
	binary.LittleEndian.PutUint32(buffer[16:20], push.FileIndex)
	addrBuffer := push.Addr.ToBytes()
	copy(buffer[20:24], addrBuffer[2:])
	copy(buffer[24:], addrBuffer[:2])
	return buffer
}
