package cluster

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

type header struct {
	pkgLen int64
	cmd    byte
	status byte
}

// Return header as a bytes buffer
func (h *header) buffer() *bytes.Buffer {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, h.pkgLen)
	//cmd
	buffer.WriteByte(h.cmd)
	//status
	buffer.WriteByte(h.status)
	return buffer
}

// Read fdfs header from conn
func (h *header) read(conn net.Conn) error {
	data := make([]byte, 10)
	if _, err := io.ReadFull(conn, data); err != nil {
		return err
	}
	buff := bytes.NewBuffer(data)
	binary.Read(buff, binary.BigEndian, &h.pkgLen)
	if h.pkgLen < 0 {
		return fmt.Errorf("wrong pkg length: %d", h.pkgLen)
	}
	h.cmd, _ = buff.ReadByte()
	h.status, _ = buff.ReadByte()
	if h.status != 0 {
		return h.statusCodeErr()
	}
	return nil
}

func (h *header) statusCodeErr() error {
	switch h.status {
	case 2:
		return errors.New("receive fileNotExist status code 2")
	case 22:
		return errors.New("receive invalidParameter status code 22")
	default:
		return fmt.Errorf("status code %d != 0", int(h.status))
	}
}
