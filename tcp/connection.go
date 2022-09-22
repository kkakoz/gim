package tcp

import (
	"github.com/kkakoz/gim"
	"github.com/kkakoz/gim/pkg/endian"
	"io"
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
	return WriteFrame(c.Conn, code, payload)
}

func (c *TcpConn) Flush() error {
	return nil
}

// WriteFrame write a frame to w
func WriteFrame(w io.Writer, code gim.OpCode, payload []byte) error {
	if err := endian.WriteUint8(w, uint8(code)); err != nil {
		return err
	}
	if err := endian.WriteBytes(w, payload); err != nil {
		return err
	}
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
