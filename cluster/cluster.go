package cluster

import (
	"math/rand"
	"strings"
	"sync"
	"time"
)

// Cluster implements a FastDFS cluster client. It manages tracker peers and multiple storage groups.
type Cluster struct {
	// user defined cluster name
	name string

	// base tracker config. Currently not used.
	trackerBaseConfig TrackerConfig

	// storage base config. If cannot find specified config to create a storage,
	// this config will be used
	storageBaseConfig StorageConfig

	// storage group map, use group name as key
	storageGroups sync.Map

	// tracker peers belong to the cluster
	trackerPeers []*Tracker

	mtx sync.RWMutex
}

// New create a cluster with specified name.
func New(name string) *Cluster {
	return &Cluster{
		name:         name,
		trackerPeers: make([]*Tracker, 0, 1),
	}
}

// DefaultCluster is created to let user call methods Download, Upload, Delete and etc through package name
var DefaultCluster = New("default")

// Init initialize cluster trackers and set up tracker and storage base config.
//
// Init is a wrapper of DefaultCluster.Init
func Init(trackerAddress []string, trackerBaseConfig TrackerConfig, storageBaseConfig StorageConfig) error {
	return DefaultCluster.Init(trackerAddress, trackerBaseConfig, storageBaseConfig)
}

// Init initialize cluster trackers and set up tracker and storage base config.
func (c *Cluster) Init(trackerAddress []string, trackerBaseConfig TrackerConfig, storageBaseConfig StorageConfig) error {
	for _, addr := range trackerAddress {
		t, err := NewTracker(addr, trackerBaseConfig)
		if err != nil {
			return c.wrapError(err)
		}
		c.AddTracker(t)
	}
	c.trackerBaseConfig = trackerBaseConfig
	c.storageBaseConfig = storageBaseConfig
	return nil
}

// AddTracker append a tracker to cluster tracker peers
func (c *Cluster) AddTracker(t *Tracker) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.trackerPeers = append(c.trackerPeers, t)
}

// AddStorageGroup add a storage group to cluster storage group map.
func (c *Cluster) AddStorageGroup(sg *StorageGroup) {
	c.storageGroups.Store(sg.groupName, sg)
}

// Append bytes to the end of the file.
//
// Append is a wrapper of DefaultCluster.Append.
func Append(b []byte, fid string) error {
	return DefaultCluster.Append(b, fid)
}

// Append bytes to the end of the file.
func (c *Cluster) Append(b []byte, fid string) error {
	//split file id to two parts: group name and file name
	group, filename, err := c.splitFid(fid)
	if err != nil {
		return err
	}
	//query a upload server from tracker
	storeInfo, err := c.Tracker().QueryUpdateStorage(group, filename)
	if err != nil {
		return c.wrapError(err)
	}
	//get a storage client from storage map, if not exist, create a new storage client
	s, err := c.Storage(storeInfo)
	if err != nil {
		return err
	}
	return s.Append(b, filename)
}

// Delete the file in this cluster.
//
// Delete is a wrapper of DefaultCluster.Delete.
func Delete(fid string) error {
	return DefaultCluster.Delete(fid)
}

// Delete the file in this cluster.
func (c *Cluster) Delete(fid string) *Error {
	group, filename, err := c.splitFid(fid)
	if err != nil {
		return err
	}
	//query a upload server from tracker
	info, err := c.Tracker().QueryUpdateStorage(group, filename)
	if err != nil {
		return c.wrapError(err)
	}
	//get a storage client from storage map, if not exist, create a new storage client
	s, err := c.Storage(info)
	if err != nil {
		return err
	}
	return s.Delete(filename)
}

// Download the whole file.
//
// Download is a wrapper of DefaultCluster.Download.
func Download(fid string) ([]byte, error) {
	return DefaultCluster.Download(fid)
}

// Download the whole file.
func (c *Cluster) Download(fid string) ([]byte, error) {
	return c.DownloadFromOffset(fid, 0, 0)
}

// DownloadFromOffset download length bytes from offset
func (c *Cluster) DownloadFromOffset(fid string, offset, length int64) ([]byte, *Error) {
	//split file id to two parts: group name and file name
	group, filename, err := c.splitFid(fid)
	if err != nil {
		return nil, err
	}
	//query a download server from tracker
	storeInfo, err := c.Tracker().QueryDownloadStorage(group, filename)
	if err != nil {
		return nil, c.wrapError(err)
	}

	//get a storage client from storage map, if not exist, create a new storage client
	s, err := c.Storage(storeInfo)
	if err != nil {
		return nil, err
	}
	b, err := s.Download(filename, offset, length)
	if err != nil {
		return nil, c.wrapError(err)
	}

	return b, nil
}

// StorageGroup query a storage group from cluster storage group map.
func (c *Cluster) StorageGroup(group string) (*StorageGroup, bool) {
	v, ok := c.storageGroups.Load(group)
	if !ok {
		return nil, false
	}
	return v.(*StorageGroup), true
}

// Storage return a stored or create a new storage based on TrackerStoreInfo.
func (c *Cluster) Storage(info *TrackerStoreInfo) (*Storage, *Error) {
	sg, ok := c.StorageGroup(info.Group)
	if !ok {
		sg = NewStorageGroup(info.Group, c.storageBaseConfig)
		c.AddStorageGroup(sg)
	}
	if s, ok := sg.Storage(info.Address); ok {
		return s, nil
	}
	s, err := NewStorage(info.Address, info.Group, sg.BaseConfig())
	if err != nil {
		return nil, c.wrapError(err)
	}
	sg.Add(s)
	return s, nil
}

// Tracker select a random tracker from cluster tracker peers.
func (c *Cluster) Tracker() *Tracker {
	c.mtx.RLock()
	defer c.mtx.RUnlock()

	rand.Seed(time.Now().Unix())
	return c.trackerPeers[rand.Intn(len(c.trackerPeers))]
}

// UpdateStorageGroup update existing group's storage config or create a new group use the config.
//
// UpdateStorageGroup is a wrapper of DefaultCluster.UpdateStorageGroup.
func UpdateStorageGroup(group string, config StorageConfig) {
	DefaultCluster.UpdateStorageGroup(group, config)
}

// UpdateStorageGroup update existing group's storage config or create a new group use the config.
func (c *Cluster) UpdateStorageGroup(group string, config StorageConfig) {
	sg, ok := c.StorageGroup(group)
	if ok {
		sg.Update(config)
	} else {
		c.AddStorageGroup(NewStorageGroup(group, config))
	}
}

// UpdateTracker update all tracker peers config.
//
// UpdateTracker is a wrapper of DefaultCluster.UpdateTracker.
func UpdateTracker(config TrackerConfig) {
	DefaultCluster.UpdateTracker(config)
}

// UpdateTracker update all tracker peers config.
func (c *Cluster) UpdateTracker(config TrackerConfig) {
	c.mtx.RLock()
	defer c.mtx.RUnlock()

	for _, t := range c.trackerPeers {
		t.Update(config)
	}
}

func (c *Cluster) upload(b []byte, group, ext string, allowAppend bool) (string, *Error) {
	//query a upload server from tracker
	t := c.Tracker()
	info, err := t.QueryUploadStorage(group)
	if err != nil {
		return "", c.wrapError(err)
	}
	//get a storage client from storage map, if not exist, create a new storage client
	s, err := c.Storage(info)
	if err != nil {
		return "", err
	}
	fid, err := s.Upload(b, info.PathIndex, ext, allowAppend)
	return fid, c.wrapError(err)
}

// Upload a file to the group with specified extension name.
// The upload cannot be appended bytes to.
// If you need to append bytes later, use UploadAppender method instead.
//
// Upload is a wrapper of DefaultCluster.Upload.
func Upload(b []byte, group, ext string) (string, error) {
	return DefaultCluster.Upload(b, group, ext)
}

// Upload a file to the group with specified extension name.
// The upload cannot be appended bytes to.
// If you need to append bytes later, use UploadAppender method instead.
func (c *Cluster) Upload(b []byte, group, ext string) (string, error) {
	return c.upload(b, group, ext, false)
}

// uploadAppender upload a file to the group with specified extension name.
// The uploaded file can be appended bytes to.
//
// UploadAppender is a wrapper of DefaultCluster.UploadAppender.
func UploadAppender(b []byte, group, ext string) (string, error) {
	return DefaultCluster.UploadAppender(b, group, ext)
}

// uploadAppender upload a file to the group with specified extension name.
// The uploaded file can be appended bytes to.
func (c *Cluster) UploadAppender(b []byte, group, ext string) (string, error) {
	return c.upload(b, group, ext, true)
}

// UploadSlave upload as a slave file of master, slave file id is {master}{suffix}.{ext}.
//
// UploadSlave is a wrapper of DefaultCluster.UploadSlave.
func UploadSlave(b []byte, master, suffix, ext string) (string, error) {
	return DefaultCluster.UploadSlave(b, master, suffix, ext)
}

// UploadSlave upload as a slave file of master, slave file id is {master}{suffix}.{ext}
func (c *Cluster) UploadSlave(b []byte, master, suffix, ext string) (string, error) {
	group, filename, err := c.splitFid(master)
	if err != nil {
		return "", err
	}
	//query a upload server from tracker
	info, err := c.Tracker().QueryUpdateStorage(group, filename)
	if err != nil {
		return "", c.wrapError(err)
	}
	//get a storage client from storage map, if not exist, create a new storage client
	s, err := c.Storage(info)
	if err != nil {
		return "", err
	}
	fid, err := s.UploadSlave(b, master, suffix, ext)
	return fid, c.wrapError(err)
}

// splitFid split file id to group name and file name
func (c *Cluster) splitFid(fid string) (string, string, *Error) {
	s := strings.SplitN(fid, "/", 2)
	if len(s) < 2 {
		return "", "", c.wrapError(wrongFidErr(fid))
	}

	return s[0], s[1], nil
}

// wrapError wrap cluster name to error
func (c *Cluster) wrapError(err *Error) *Error {
	if err == nil {
		return nil
	}
	return err.Wrap(c.name)
}
