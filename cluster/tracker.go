package cluster

import (
	"encoding/binary"
	"fmt"
	"github.com/giantpoplar/pool"
)

// Tracker implements a client to access a FastDFS tracker node
type Tracker struct {
	// Address is with format host:port
	address string

	// Connection pool
	pool pool.Pool
}

// NewTracker creates a client to access tracker node. It uses a blocking pool
// to manage net connection. So created connection can be precisely controlled.
// Note that if some config items are not set, default config will be used.
func NewTracker(address string, config TrackerConfig) (*Tracker, *Error) {
	t := &Tracker{address: address}
	p, err := pool.NewBlockingPool(address, config.PoolConfig, nil)
	if err != nil {
		return nil, t.wrapError(createPoolErr(err))
	}
	t.pool = p
	return t, nil
}

// TrackerStoreInfo is tracker return storage info for uploading, downloading and etc.
type TrackerStoreInfo struct {
	// Address
	Address string

	// Group name
	Group string

	// PathIndex is storage volume index
	PathIndex byte
}

// cast receive bytes to TrackerStoreInfo
func (tsi *TrackerStoreInfo) cast(recv []byte, containPath bool) *Error {
	if containPath {
		if len(recv) != TRACKER_QUERY_STORAGE_STORE_BODY_LEN {
			return unexpectedPkgLenErr(len(recv), TRACKER_QUERY_STORAGE_STORE_BODY_LEN)
		}
		tsi.PathIndex = recv[39]
	} else {
		if len(recv) != TRACKER_QUERY_STORAGE_FETCH_BODY_LEN {
			return unexpectedPkgLenErr(len(recv), TRACKER_QUERY_STORAGE_FETCH_BODY_LEN)
		}
	}

	tsi.Group = stripString(string(recv[:16]))
	ip := stripString(string(recv[16:31]))
	port := binary.BigEndian.Uint64(recv[31:39])
	tsi.Address = fmt.Sprintf("%s:%d", ip, port)
	return nil
}

// QueryUploadStorage query group upload storage info for update
func (t *Tracker) QueryUploadStorage(group string) (*TrackerStoreInfo, *Error) {
	//get a connection from pool
	conn, e := t.pool.Get()
	if e != nil {
		return nil, t.wrapError(getConnErr(e))
	}
	defer conn.Close()

	h := &header{
		pkgLen: int64(FDFS_GROUP_NAME_MAX_LEN),
		cmd:    TRACKER_PROTO_CMD_SERVICE_QUERY_STORE_WITH_GROUP_ONE,
	}
	buffer := h.buffer()
	//16 bit groupName
	buffer.WriteString(fixString(group, FDFS_GROUP_NAME_MAX_LEN))

	r := request{
		c:      conn,
		header: buffer.Bytes(),
	}
	recv, err := r.do()
	if err != nil {
		return nil, err
	}

	info := &TrackerStoreInfo{}
	if err := info.cast(recv, true); err != nil {
		return nil, t.wrapError(err)
	}
	return info, nil
}

// QueryUpdateStorage query storage info for update actions like delete and append
func (t *Tracker) QueryUpdateStorage(group, filename string) (*TrackerStoreInfo, *Error) {
	return t.queryFileStorage(group, filename, TRACKER_PROTO_CMD_SERVICE_QUERY_UPDATE)
}

// QueryDownloadStorage query storage info for download
func (t *Tracker) QueryDownloadStorage(group, filename string) (*TrackerStoreInfo, *Error) {
	return t.queryFileStorage(group, filename, TRACKER_PROTO_CMD_SERVICE_QUERY_FETCH_ONE)
}

// Query stroage info using filename with specific command
func (t *Tracker) queryFileStorage(group, filename string, cmd byte) (*TrackerStoreInfo, *Error) {
	//get a connection from pool
	conn, e := t.pool.Get()
	if e != nil {
		return nil, t.wrapError(getConnErr(e))
	}
	defer conn.Close()

	h := &header{
		pkgLen: int64(FDFS_GROUP_NAME_MAX_LEN + len(filename)),
		cmd:    cmd,
	}
	buffer := h.buffer()
	//16 bit groupName
	buffer.WriteString(fixString(group, FDFS_GROUP_NAME_MAX_LEN))
	// fileName
	buffer.WriteString(filename)

	r := request{
		c:      conn,
		header: buffer.Bytes(),
	}
	recv, err := r.do()
	if err != nil {
		return nil, err
	}

	info := &TrackerStoreInfo{}
	if err = info.cast(recv, false); err != nil {
		return nil, t.wrapError(err)
	}
	return info, nil
}

// Update tracker pool config
func (t *Tracker) Update(config TrackerConfig) {
	t.pool.Update(config.PoolConfig)
}

// wrapError wrap tracker relevant header to the error name
func (t *Tracker) wrapError(err *Error) *Error {
	if err == nil {
		return err
	}
	return err.Wrap("Tracker:" + t.address)
}
