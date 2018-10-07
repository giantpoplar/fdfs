package cluster

import (
	"encoding/binary"
	"path/filepath"

	"fmt"
	"github.com/giantpoplar/pool"
	"sync"
)

// Storage implements a client to access a FastDFS storage node.
type Storage struct {
	// address is with format host:port
	address string

	config StorageConfig

	// group name
	group string

	// conntection pool
	pool pool.Pool

	mtx sync.RWMutex
}

// NewStorage creates a client to access storage node. It uses a blocking pool
// to manage net connection. So created connection can be precisely controlled.
// Note that if some config items are not set, default config will be used.
func NewStorage(address, group string, config StorageConfig) (*Storage, *Error) {
	s := &Storage{
		address: address,
		group:   group,
		config:  defaultStorageConfig.merge(config),
	}
	p, err := pool.NewBlockingPool(address, s.config.PoolConfig, nil)
	if err != nil {
		return nil, s.wrapError(createPoolErr(err))
	}
	s.pool = p
	return s, nil
}

// Append bytes to the file
func (s *Storage) Append(b []byte, filename string) *Error {
	//get a connetion from pool
	conn, e := s.pool.Get()
	if e != nil {
		return s.wrapError(getConnErr(e))
	}
	defer conn.Close()

	h := &header{
		pkgLen: int64(16 + len(filename) + len(b)),
		cmd:    STORAGE_PROTO_CMD_APPEND_FILE,
	}
	buffer := h.buffer()
	//8 bytes: appender filename length
	binary.Write(buffer, binary.BigEndian, int64(len(filename)))
	//8 bytes: file size
	binary.Write(buffer, binary.BigEndian, int64(len(b)))
	//appender file name
	buffer.WriteString(filename)

	req := request{
		c:         conn,
		header:    buffer.Bytes(),
		body:      b,
		respLimit: 130,
	}
	_, err := req.do()
	return s.wrapError(err)
}

// Delete file
func (s *Storage) Delete(filename string) *Error {
	//get a connetion from pool
	conn, e := s.pool.Get()
	if e != nil {
		return s.wrapError(getConnErr(e))
	}
	defer conn.Close()

	h := &header{
		pkgLen: int64(FDFS_GROUP_NAME_MAX_LEN + len(filename)),
		cmd:    STORAGE_PROTO_CMD_DELETE_FILE,
	}
	buffer := h.buffer()
	//16 bit groupName
	buffer.WriteString(fixString(s.group, FDFS_GROUP_NAME_MAX_LEN))
	// fileNameLen bit fileName
	buffer.WriteString(filename)

	req := request{c: conn, header: buffer.Bytes()}
	_, err := req.do()
	return s.wrapError(err)
}

func (s *Storage) downloadSizeLimit() int64 {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return s.config.DownloadSizeLimit
}

func (s *Storage) setDownloadSizeLimit(limit int64) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.config.DownloadSizeLimit = limit
}

// Download length bytes of file from offset
func (s *Storage) Download(filename string, offset, length int64) ([]byte, *Error) {
	//get a connetion from pool
	conn, e := s.pool.Get()
	if e != nil {
		return nil, s.wrapError(getConnErr(e))
	}
	defer conn.Close()

	h := &header{
		pkgLen: int64(32 + len(filename)),
		cmd:    STORAGE_PROTO_CMD_DOWNLOAD_FILE,
	}
	buffer := h.buffer()
	// Request: file_offset(8)  download_bytes(8)  group_name(16)  file_name(n)
	// offset
	binary.Write(buffer, binary.BigEndian, offset)
	// download bytes
	binary.Write(buffer, binary.BigEndian, length)
	// 16 bit groupName
	buffer.WriteString(fixString(s.group, FDFS_GROUP_NAME_MAX_LEN))
	// fileName
	buffer.WriteString(filename)

	req := request{
		c:         conn,
		header:    buffer.Bytes(),
		respLimit: s.downloadSizeLimit(),
	}
	recv, err := req.do()
	return recv, s.wrapError(err)
}

// Update storage config with new one
func (s *Storage) Update(config StorageConfig) {
	if config.DownloadSizeLimit > 0 {
		s.setDownloadSizeLimit(config.DownloadSizeLimit)
	}
	s.pool.Update(config.PoolConfig)
}

// Upload a file to the storage path.
func (s *Storage) Upload(b []byte, pathIndex byte, ext string, allowAppend bool) (string, *Error) {
	//get a connetion from pool
	conn, e := s.pool.Get()
	if e != nil {
		return "", s.wrapError(getConnErr(e))
	}
	defer conn.Close()

	cmd := STORAGE_PROTO_CMD_UPLOAD_FILE
	if allowAppend {
		cmd = STORAGE_PROTO_CMD_UPLOAD_APPENDER_FILE
	}

	h := &header{
		pkgLen: int64(15 + len(b)),
		cmd:    byte(cmd),
	}
	buffer := h.buffer()
	//store_path_index
	buffer.WriteByte(pathIndex)
	// file size
	binary.Write(buffer, binary.BigEndian, int64(len(b)))
	// 6 bit fileExtName
	buffer.WriteString(fixString(ext, FDFS_FILE_EXT_NAME_MAX_LEN))

	req := request{c: conn, header: buffer.Bytes(), body: b, respLimit: 130}
	recv, err := req.do()
	if err != nil {
		return "", s.wrapError(err)
	}
	return s.parseFid(recv)
}

// Upload a slave file. Slave file id is {master}{suffix}.{ext}
func (s *Storage) UploadSlave(b []byte, master, suffix, ext string) (string, *Error) {
	//get a connetion from pool
	conn, e := s.pool.Get()
	if e != nil {
		return "", s.wrapError(getConnErr(e))
	}
	defer conn.Close()

	h := &header{
		//master_len(8) file_size(8) prefix_name(16) file_ext_name(6) master_name(master_filename_len)
		pkgLen: int64(38 + len(master) + len(b)),
		cmd:    STORAGE_PROTO_CMD_UPLOAD_FILE,
	}
	buffer := h.buffer()

	// master file name len
	binary.Write(buffer, binary.BigEndian, int64(len(master)))
	// file size
	binary.Write(buffer, binary.BigEndian, int64(len(b)))
	// 16 bit prefixName
	buffer.WriteString(fixString(suffix, FDFS_FILE_PREFIX_MAX_LEN))
	// 6 bit fileExtName
	buffer.WriteString(fixString(ext, FDFS_FILE_EXT_NAME_MAX_LEN))
	// master_file_name
	buffer.WriteString(master)

	req := request{
		c:         conn,
		header:    buffer.Bytes(),
		body:      b,
		respLimit: 130,
	}
	recv, err := req.do()
	if err != nil {
		return "", s.wrapError(err)
	}
	return s.parseFid(recv)
}

// parseFid parse receive bytes to group and name
func (s *Storage) parseFid(recv []byte) (string, *Error) {
	// #recv_fmt |-group_name(16)-filename|
	if len(recv) < FDFS_GROUP_NAME_MAX_LEN {
		return "", s.wrapError(unexpectedPkgLenErr(len(recv), FDFS_GROUP_NAME_MAX_LEN))
	}
	group := stripString(string(recv[0:FDFS_GROUP_NAME_MAX_LEN]))
	name := string(recv[FDFS_GROUP_NAME_MAX_LEN:])
	return filepath.Join(group, name), nil
}

// wrapError wrap storage group and address to the error
func (s *Storage) wrapError(err *Error) *Error {
	if err == nil {
		return err
	}
	return err.Wrap(fmt.Sprintf("Storage_%s:%s", s.group, s.address))
}
