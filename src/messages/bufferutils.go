package messages

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func writeStringLE(buff []byte, str string) error {
	minLen := min(len(buff), len(str))
	buffer := bytes.NewBufferString(str)
	for i := 0; i < minLen; i++ {
		err := binary.Write(buffer, binary.LittleEndian, str[i])
		if err != nil {
			return err
		}
	}
	copy(buff, buffer.Bytes())
	return nil
}

func readStringLE(buff []byte) (string, error) {
	reader := bytes.NewReader(buff)
	writer := make([]byte, len(buff))
	err := binary.Read(reader, binary.LittleEndian, writer)
	if err != nil {
		return nil, err
	}
	return string(writer), nil
}

func min(x, y int) (min int) {
	if x > y {
		min = y
	} else {
		min = x
	}
	return
}
