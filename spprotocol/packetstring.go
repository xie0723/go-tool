package spprotocol

import (
	"bytes"
	"encoding/binary"
)

//
// 生成一个string返回报文
//

func GenBytePacket(header string, context []byte) []byte {
	var buffer_client bytes.Buffer
	buffer_client.WriteString(header)
	var kindBytes = make([]byte, 2)
	binary.BigEndian.PutUint16(kindBytes, uint16(1))
	buffer_client.Write(kindBytes)
	var contextLengthBytes = make([]byte, 4)
	binary.BigEndian.PutUint32(contextLengthBytes, uint32(len(context)))
	buffer_client.Write(contextLengthBytes)
	buffer_client.Write(context)
	return buffer_client.Bytes()
}
