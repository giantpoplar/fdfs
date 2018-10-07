/*
Package cluster is a pure golang client for a single FastDFS cluster.

A cluster is defined as a tracker peer group with multiple storage groups. If you have several
independent clusters for particular purpose, use package fdfs instead.

Initialize a cluster is simple:

	err := cluster.Init([]string{"127.0.0.1:22122"}, cluster.TrackerConfig{}, cluster.StorageConfig{})

This will initialize default cluster with default tracker and storage config. You can specify your
own config:

	myTrackerConfig := cluster.TrackerConfig{
		PoolConfig: pool.Config{
			CacheMethod: pool.FIFO,        // connection pool organized method, first come in conn will be out firstly to use.
			InitCap:     1,                // Conn will be created in initialized.
			MaxCap:      3,                // Max conn can be created.
			IdleTimeout: 30 * time.Second, // Conn whose the idle time exceeds this time will be closed
			WaitTimeout: 3 * time.Second,  // Time to wait a idle conn.
			IOTimeout:   30 * time.Second, // Conn Read and Write timeout.
			DialTimeout: 30 * time.Second, // net.Dial timeout.
		},
	}
	myStorageConfig := cluster.StorageConfig{
		// Limit download max size 128M. Find exceed this limit will return error directly.
		// Set this limit is meaningful. For example, download a 4GB file will cause a shock
		// to both client and server. So, we need to fast fail.
		DownloadSizeLimit: 128 * 1024 * 1024,
		PoolConfig: pool.Config{
			CacheMethod: pool.FILO,        // connection pool organized method, first come in conn will be out last use.
			InitCap:     2,                // Conn will be created in initialized.
			MaxCap:      8,                // Max conn can be created.
			IdleTimeout: 30 * time.Second, // Conn whose the idle time exceeds this time will be closed
			WaitTimeout: 3 * time.Second,  // Time to wait a idle conn.
			IOTimeout:   30 * time.Second, // Conn Read and Write timeout.
			DialTimeout: 30 * time.Second, // net.Dial timeout.
		},
	}
	err := cluster.Init([]string{"127.0.0.1:22122"}, myTrackerConfig, myStorageConfig)

You can also hot update tracker and storage config:

	cluster.UpdateTracker(your_tracker_config)

will update all trackers config.

	cluster.UpdateStorageGroup("g1", your_storage_config)

will update all storage belong to 'g1'.

After initialization, you can do Upload, Download, Delete and etc actions:

	fid, err := cluster.Upload("g1", "jpg", b)
	...
	b, err := cluster.Download("g1/M01/DE/79/CgIG6VuXIoeAbiwbAAAIIRe5FG4412.jpg")
	...
	err := cluster.Delete("g1/M01/DE/79/CgIG6VuXIoeAbiwbAAAIIRe5FG4412.jpg")

 */
package cluster
