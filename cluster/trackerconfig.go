package cluster

import (
	"github.com/giantpoplar/pool"
	"time"
)

// TrackerConfig defines necessary parameters to create a tracker.
type TrackerConfig struct {
	// Tracker connection pool config
	PoolConfig pool.Config
}

var defaultTrackerConfig = TrackerConfig{
	PoolConfig: pool.Config{
		CacheMethod: pool.FIFO,
		IdleTimeout: 3 * time.Second,
		WaitTimeout: 3 * time.Second,
		IOTimeout:   30 * time.Second,
		DialTimeout: 30 * time.Second,
		InitCap:     1,
		MaxCap:      3,
	},
}

// merge new config to old one.
func (tc *TrackerConfig) merge(new TrackerConfig) TrackerConfig {
	result := *tc
	result.PoolConfig, _ = result.PoolConfig.Merge(new.PoolConfig)
	return result
}
