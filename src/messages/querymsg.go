package messages

import (
	"encoding/binary"
	"fmt"
)

type QueryMsg struct {
	MinSpeed    uint16
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
	query.MinSpeed = binary.LittleEndian.Uint16(buffer[:2])

	queryBuffer := buffer[2:]
	nullIdx := findNullByte(queryBuffer)
	if nullIdx == -1 {
		return fmt.Errorf("Input buffer didn't have null terminating search query string")
	}
	query.SearchQuery = ReadStringLE(queryBuffer[:nullIdx]) // cut off null byte
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
	bufferLen := 3 + len(query.SearchQuery) // 2 bytes for MinSpeed, len(query.SearchQuery) bytes for query, 1 byte for null terminating char
	buffer := make([]byte, bufferLen)
	binary.LittleEndian.PutUint16(buffer[:2], query.MinSpeed)
	WriteStringLE(buffer[2:], query.SearchQuery)
	buffer[bufferLen-1] = 0x00
	return buffer
}
