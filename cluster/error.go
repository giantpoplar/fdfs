package cluster

import "fmt"

type Error struct {
	name   string
	detail error
}

func NewError(name string, err error) *Error {
	return &Error{name, err}
}

func (err *Error) Error() string {
	return fmt.Sprintf("%s:%s", err.name, err.detail)
}

func (err *Error) Name() string {
	return err.name
}

func (err *Error) Wrap(name string) *Error {
	err.name = fmt.Sprintf("%s.%s", name, err.name)
	return err
}

func createPoolErr(err error) *Error {
	return NewError("CreatePoolErr", err)
}

func getConnErr(err error) *Error {
	return NewError("GetConnFromPoolErr", err)
}

func unexpectedPkgLenErr(receive, expect int) *Error {
	return NewError(" UnexpectedLenErr", fmt.Errorf("received pkg length %d != expected %d", receive, expect))
}

func wrongFidErr(fid string) *Error {
	return NewError("WrongFidErr", fmt.Errorf("fid is not with format group/filename: %s", fid))
}
