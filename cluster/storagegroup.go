package cluster

import (
	"sync"
)

// StorageGroup manage storage belong to same FastDFS group.
type StorageGroup struct {
	// group name
	groupName string

	// group shared base storage config
	base StorageConfig

	// storage map, key is storage address
	storageMap map[string]*Storage

	mtx sync.RWMutex
}

// NewStorageGroup create a storageGroup to manage storage belong to same group.
func NewStorageGroup(groupName string, base StorageConfig) *StorageGroup {
	return &StorageGroup{
		base:       base,
		groupName:  groupName,
		storageMap: make(map[string]*Storage),
	}
}

// Add add a storage to group
func (sg *StorageGroup) Add(s *Storage) *StorageGroup {
	sg.mtx.Lock()
	defer sg.mtx.Unlock()

	sg.storageMap[s.address] = s
	return sg
}

// BaseConfig return group shared base storage config.
func (sg *StorageGroup) BaseConfig() StorageConfig {
	sg.mtx.RLock()
	defer sg.mtx.RLock()

	return sg.base
}

// Storage return query result of storage map
func (sg *StorageGroup) Storage(addr string) (*Storage, bool) {
	sg.mtx.RLock()
	defer sg.mtx.RUnlock()

	s, ok := sg.storageMap[addr]
	if !ok {
		return nil, false
	}
	return s, true
}

// Update group all storage config
func (sg *StorageGroup) Update(config StorageConfig) {
	sg.mtx.RLock()
	for _, s := range sg.storageMap {
		s.Update(config)
	}
	sg.mtx.RUnlock()
	sg.mtx.Lock()
	defer sg.mtx.Unlock()
	sg.base = config

}
