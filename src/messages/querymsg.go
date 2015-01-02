package message

import (
	"encoding/binary"
	"fmt"
)

type QueryMsg struct {
	MinSpeed    uint32
	SearchQuery string
}

func findNullByte(buffer []byte) int {
	// Reverse iteration...chances are null char will be towards the end of the buffer
	for i := len(buffer) - 1; i >= 0; i-- {
		if buffer[i] == 0x00 {
			return i
		}
	}
	return -1
}

func parseQueryBytes(buffer []byte, query *QueryMsg) error {
	if len(buffer) <= 3 {
		return fmt.Errorf("Expected buffer to be of length > 3. Got buffer of length %d", len(buffer))
	}
	query.MinSpeed = binary.LittleEndian.Uint32(buffer[:2])

	nullIdx := findNullByte(buffer[2:])
	if nullIdx == -1 {
		return fmt.Errorf("Input buffer didn't have null terminating search query string")
	}
	queryBuffer := buffer[:nullIdx]
	query.SearchQuery = string(queryBuffer)
	return nil
}

func ParseQueryBytes(buffer []byte) (*QueryMsg, error) {
	query := new(QueryMsg)
	err := parseQueryBytes(buffer, query)
	return query, err
}

func (query *QueryMsg) ParseBytes(buffer []byte) error {
	err := parseQueryBytes(buffer, query)
	return err
}

func (query *QueryMsg) ToBytes() []byte {
	bufferLen := 3 + len(query.SearchQuery) // 2 bytes for MinSpeed, len(query.SearchQuery) bytes for query, 1 bytes for null terminating char
	buffer := make([]byte, bufferLen)
	binary.LittleEndian.PutUint32(buffer[:2], query.MinSpeed)
	copy(buffer[2:], query.SearchQuery)
	buffer[bufferLen-1] = 0x00
	return buffer
}
