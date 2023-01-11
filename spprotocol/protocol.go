package spprotocol

//
// spprotocol  simple private protocol  简单私有协议
//
import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

const CONTENT_SIZE = 4 // 4个字节用来标示内容长度
const KIND_SIZE = 2    // 存储消息类型
const KIND_END = 1     // 最后一条消息
const KIND_NEXT = 2    // 还有下一条小时

type ProtocolCallBackFunc func(conn net.Conn, content []byte) error
type Buffer struct {
	conn      net.Conn
	header    string
	buf       []byte
	bufLength int
	callBack  ProtocolCallBackFunc
	start     int
	end       int
}

func NewBuffer(conn net.Conn, header string, len int, callBack ProtocolCallBackFunc) *Buffer {
	buf := make([]byte, len)
	return &Buffer{conn, header, buf, len, callBack, 0, 0}
}

// grow 将有效的字节前移
func (b *Buffer) grow() {
	if b.start == 0 {
		return
	}
	copy(b.buf, b.buf[b.start:b.end])
	b.end -= b.start
	b.start = 0
}

// 缓存区中有效字节长度
func (b *Buffer) len() int {
	return b.end - b.start
}

// 返回n个字节，而不产生移位
func (b *Buffer) seek(n int) ([]byte, error) {
	if b.end-b.start >= n {
		buf := b.buf[b.start : b.start+n]
		return buf, nil
	}
	return nil, errors.New("not enough")
}

// 舍弃offset个字节，读取n个字节
func (b *Buffer) read(offset, n int) []byte {
	b.start += offset
	buf := b.buf[b.start : b.start+n]
	b.start += n
	return buf
}

// 从reader里面读取数据，如果reader阻塞，会发生阻塞
func (b *Buffer) readFromReader() error {
	if b.end == b.bufLength {
		return fmt.Errorf("content length is large the max buf length %d", b.bufLength)
	}
	n, err := b.conn.Read(b.buf[b.end:])
	if err != nil {
		return err
	}
	//time.Sleep(1 * time.Second) // 便于观察这里sleep了一下
	b.end += n
	return nil
}

// 处理 conn
func (buffer *Buffer) Handle() (err error) {
	for {
		buffer.grow() // 移动数据
		// 读数据拼接到定额缓存后面
		if err = buffer.readFromReader(); err != nil {
			return err // 读取缓存信息失败，可能是对方关闭连接
		}
		HEADER_LENG := CONTENT_SIZE + KIND_SIZE + len(buffer.header)
		var headBuf []byte
		if headBuf, err = buffer.seek(HEADER_LENG); err != nil {
			// 缓存中消息不够一个消息头的长度， 跳出去继续读
			continue
		}
		if string(headBuf[:len(buffer.header)]) != buffer.header { // 判断消息头正确性
			return errors.New("massage head is incorrect, expect " + buffer.header + " but " + string(headBuf[:len(buffer.header)]))
		}
		kind := int(binary.BigEndian.Uint16(headBuf[len(buffer.header) : len(buffer.header)+KIND_SIZE]))
		contentSize := int(binary.BigEndian.Uint32(headBuf[len(buffer.header)+KIND_SIZE:]))
		if buffer.len() >= contentSize-HEADER_LENG {
			// 满足一个消息体
			contentBuf := buffer.read(HEADER_LENG, contentSize) // 把消息读出来，把start往后移
			if err = buffer.callBack(buffer.conn, contentBuf); err != nil {
				// 回调函数出错，结束
				return
			}
			if kind == KIND_NEXT {
				// 还有下一条消息
				continue
			}
			return nil
		}
	}
}
