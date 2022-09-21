package tcp

import (
	"github.com/kkakoz/gim"
	"github.com/kkakoz/gim/pkg/endian"
	"net"
)

type TcpConn struct {
	net.Conn
}

func NewConn(conn net.Conn) *TcpConn {
	return &TcpConn{
		Conn: conn,
	}
}

func (c *TcpConn) ReadFrame() (gim.Frame, error) {
	opcode, err := endian.ReadUint8(c.Conn)
	if err != nil {
		return nil, err
	}
	data, err := endian.ReadBytes(c.Conn)
	if err != nil {
		return nil, err
	}
	return &Frame{OpCode: gim.OpCode(opcode), Payload: data}, nil
}

func (c *TcpConn) WriteFrame(code gim.OpCode, payload []byte) error {
	err := endian.WriteUint8(c.Conn, uint8(code))
	if err != nil {
		return err
	}
	return endian.WriteShortBytes(c.Conn, payload)
}

func (c *TcpConn) Flush() error {
	return nil
}

type Frame struct {
	OpCode  gim.OpCode
	Payload []byte
}

func (f *Frame) SetOpCode(code gim.OpCode) {
	f.OpCode = code
}

func (f *Frame) GetOpCode() gim.OpCode {
	return f.OpCode
}

func (f *Frame) SetPayload(payload []byte) {
	f.Payload = payload
}

func (f *Frame) GetPayload() []byte {
	return f.Payload
}
