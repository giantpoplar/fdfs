/*
Package fdfs is a pure golang client for FastDFS https://github.com/happyfish100/fastdfs.

It support multiple independent fdfs clusters. A cluster is defined as a tracker peer group
with multiple storage groups. If you have only one cluster, subpackage cluster is more convenient
to use.

If you have two independent FastDFS clusters, named as 'fdfs_cluster1' and 'fdfs_cluster2',
tracker addresses of cluster 1 are 10.15.25.46:22122 and  10.15.25.47:22122, cluster 2 are
10.28.89.12:22122 and 10.28.89.13:22122, you can initialize client as follows:

	c := cluster.New("fdfs_cluster1")
	err := c.Init(
		[]string{"10.15.25.46:22122", "10.15.25.47:22122"},
		cluster.TrackerConfig{},
		cluster.StorageConfig{},
	)
	if err != nil {
		// handle error
	}
	fdfs.AddCluster(c)
	c = cluster.New("fdfs_cluster2")
	err = c.Init(
		[]string{"10.28.89.12:22122", "10.28.89.13:22122"},
		cluster.TrackerConfig{},
		cluster.StorageConfig{},
	)
	if err != nil {
		// handler error
	}
	fdfs.AddCluster(c)

You can specify tracker and storage config independently. Please visit subpackage cluster for detail.

Then, you can do Upload, Download, Delete and etc actions to a cluster:

	fid, err := fdfs.Upload("fdfs_cluster1", "g1", "jpg", b)
	...
	b, err := fdfs.Download("fdfs_cluster1", "g1/M01/DE/79/CgIG6VuXIoeAbiwbAAAIIRe5FG4412.jpg")
	...
	err := fdfs.Delete("fdfs_cluster1", "g1/M01/DE/79/CgIG6VuXIoeAbiwbAAAIIRe5FG4412.jpg")


Client also support hot update tracker and storage group config:

	fdfs.UpdateTracker("fdfs_cluter1", your_tracker_config)

This will update cluster 1 trackers: 10.15.25.46, 10.15.25.47.

	fdfs.UpdateStorageGroup("fdfs_cluster1", "g1", your_storage_config)

This will update all cluster 1 storage belong to 'g1'.

 */
package fdfs


