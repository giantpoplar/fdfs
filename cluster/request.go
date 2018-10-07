package cluster

import (
	"fmt"
	"github.com/giantpoplar/pool"
	"io"
)

type request struct {
	c         *pool.WrappedConn
	header    []byte
	body      []byte
	respLimit int64
}

func (r *request) do() ([]byte, *Error) {
	//send header
	if _, err := r.c.Write(r.header); err != nil {
		return nil, NewError("WriteRequestHeaderErr", err)
	}
	//send body
	if r.body != nil {
		if _, err := r.c.Write(r.body); err != nil {
			return nil, NewError("WriteRequestBodyErr", err)
		}
	}
	//receive server response
	return r.readResponse()
}

func (r *request) readResponse() ([]byte, *Error) {
	//receive response header
	h := header{}
	if err := h.read(r.c); err != nil {
		return nil, NewError("ReadResponseHeaderErr", err)
	}
	if r.respLimit > 0 && h.pkgLen > r.respLimit {
		r.c.MarkUnusable()
		return nil, NewError("WrongPkgLengthErr", fmt.Errorf("receive header pkg length %d exceed expected or limit size: %d", h.pkgLen, r.respLimit))
	}
	//receive body
	resp := make([]byte, h.pkgLen)
	if _, err := io.ReadFull(r.c, resp); err != nil {
		return nil, NewError("ReadResponseBodyErr", err)
	}
	return resp, nil
}
