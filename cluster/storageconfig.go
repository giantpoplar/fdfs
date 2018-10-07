package cluster

import (
	"github.com/giantpoplar/pool"
	"time"
)

// StorageConfig defines necessary parameters to create a storage.
type StorageConfig struct {
	// DownloadSizeLimit defines max bytes storage can download. Storage will return error directly when
	// find receive bytes exceed limit.
	// Set this limit is necessary because large file will occupy large amount of system resource and
	// FastDFS is not designed for it.
	DownloadSizeLimit int64

	// Storage connection pool config
	PoolConfig pool.Config
}

var defaultStorageConfig = StorageConfig{
	// download limit 128M
	DownloadSizeLimit: 128 * 1024 * 1024,

	PoolConfig: pool.Config{
		CacheMethod: pool.FILO,
		IdleTimeout: 100 * time.Second,
		WaitTimeout: 3 * time.Second,
		IOTimeout:   30 * time.Second,
		DialTimeout: 30 * time.Second,
		InitCap:     1,
		MaxCap:      3,
	},
}

// merge new config to old one.
func (sc *StorageConfig) merge(new StorageConfig) StorageConfig {
	result := *sc
	if new.DownloadSizeLimit > 0 {
		result.DownloadSizeLimit = new.DownloadSizeLimit
	}
	result.PoolConfig, _ = result.PoolConfig.Merge(new.PoolConfig)
	return result
}
