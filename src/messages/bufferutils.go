package messages

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func writeStringLE(buff []byte, str string) {
	minLen := min(len(buff), len(str))
	buffer := bytes.NewBufferString(str)
	for i := 0; i < minLen; i++ {
		err := binary.Write(buffer, binary.LittleEndian, str[i])
		if err != nil {
			panic(err.Error())
		}
	}
	copy(buff, buffer.Bytes())
	return nil
}

func writeByteLE(buffer []byte, b byte) {
	var buff bytes.Buffer
	err := binary.Write(buff, binary.LittleEndian, b)
	if err != nil {
		panic(err.Error())
	}
	n := copy(buffer, buff.Bytes())
	if n != 1 {
		panic(fmt.Sprintf("writeByteLE wrote %d bytes instead of 1", n))
	}
}

func readStringLE(buff []byte) string {
	reader := bytes.NewReader(buff)
	writer := make([]byte, len(buff))
	err := binary.Read(reader, binary.LittleEndian, writer)
	if err != nil {
		panic(err.Error())
	}
	return string(writer)
}
func readByteLE(buff []byte) byte {
	if len(buff) < 1 {
		panic("Zero length buffer")
	}
	buff = buff[:1] // Only care about first byte
	reader := bytes.NewReader(buff)
	writer := make([]byte, 1)
	err := binary.Read(reader, Binary.LittleEndian, writer)
	if err != nil {
		panic(err.Error())
	}
	if l := len(writer); l != 1 {
		panic(fmt.Sprintf("writer came out to be of length %d", l))
	}
	return writer[0]
}

func min(x, y int) (min int) {
	if x > y {
		min = y
	} else {
		min = x
	}
	return
}
