package fdfs

import (
	"fmt"
	"github.com/giantpoplar/fdfs/cluster"
	"sync"
)

// A client is a fdfs client which manage multiple fdfs clusters. These clusters are independent of each other.
// Each FastDFS cluster manage a set of tracker clients and multiple storage groups. Each storage group manage
// multiple storage clients.
//
// If you have only one cluster, you can directly use cluster API. Also, tacker and storage
// APIs are all opened to let you use maximum free.
type Client struct {
	clusters sync.Map
}

// DefaultClient is created to let user call methods Download, Upload, Delete and etc through package name.
var DefaultClient = &Client{}

// AddCluster add a cluster to client internal cluster map with cluster name as the key.
//
// AddCluster is a wrapper of DefaultClient.AddCluster
func AddCluster(cluster *cluster.Cluster) {
	DefaultClient.AddCluster(cluster)
}

// AddCluster add a cluster to client internal cluster map with cluster name as the key.
func (c *Client) AddCluster(cluster *cluster.Cluster) {
	if cluster == nil {
		return
	}
	c.clusters.Store(cluster, cluster)
}

func unknownClusterErr(name string) *cluster.Error {
	return cluster.NewError("UnknownCluster", fmt.Errorf("cluster %s not exist", name))
}

// Append bytes to the end of the file in the cluster.
// The file must first use UploadAppender method upload.
//
// Append is a wrapper of DefaultClient.Append.
func Append(clusterName, fid string, b []byte) error {
	return DefaultClient.Append(clusterName, fid, b)
}

// Append bytes to the end of the file in the cluster.
// The file must first use UploadAppender method upload.
func (c *Client) Append(clusterName, fid string, b []byte) error {
	cluster, ok := c.Cluster(clusterName)
	if !ok {
		return unknownClusterErr(clusterName)
	}
	return cluster.Append(b, fid)
}

// Cluster load stored cluster by cluster name.
//
// Cluster is a wrapper of DefaultClient.Cluster
func Cluster(name string) (*cluster.Cluster, bool) {
	return DefaultClient.Cluster(name)
}

// Cluster load stored cluster by cluster name.
func (c *Client) Cluster(name string) (*cluster.Cluster, bool) {
	v, ok := c.clusters.Load(name)
	if !ok {
		return nil, false
	}
	return v.(*cluster.Cluster), true
}

// Delete file in the cluster.
//
// Delete is a wrapper of DefaultClient.Delete.
func Delete(clusterName, fid string) error {
	return DefaultClient.Delete(clusterName, fid)
}

// Delete file in the cluster.
func (c *Client) Delete(clusterName, fid string) error {
	cluster, ok := c.Cluster(clusterName)
	if !ok {
		return unknownClusterErr(clusterName)
	}
	return cluster.Delete(fid)
}

// Download file in the cluster.
//
// Download is a wrapper of DefaultClient.Download.
func Download(clusterName, fid string) ([]byte, error) {
	return DefaultClient.Download(clusterName, fid)
}

// Download file in the cluster.
func (c *Client) Download(clusterName, fid string) ([]byte, error) {
	cluster, ok := c.Cluster(clusterName)
	if !ok {
		return nil, unknownClusterErr(clusterName)
	}
	return cluster.Download(fid)
}

// UpdateStorageGroup update cluster storage pool config belong to same group.
//
// UpdateStorageGroup is a wrapper of DefaultClient.UpdateStorageGroup.
func UpdateStorageGroup(clusterName, group string, config cluster.StorageConfig) error {
	return DefaultClient.UpdateStorageGroup(clusterName, group, config)
}

// UpdateStorageGroup update cluster storage pool config belong to same group.
func (c *Client) UpdateStorageGroup(clusterName, group string, config cluster.StorageConfig) error {
	cluster, ok := c.Cluster(clusterName)
	if !ok {
		return unknownClusterErr(clusterName)
	}
	cluster.UpdateStorageGroup(group, config)
	return nil
}

// UpdateTracker update cluster tracker peers pool config.
//
// UpdateTracker is a wrapper of DefaultClient.UpdateTracker.
func UpdateTracker(clusterName string, config cluster.TrackerConfig) error {
	return DefaultClient.UpdateTracker(clusterName, config)
}

// UpdateTracker update cluster tracker peers pool config.
func (c *Client) UpdateTracker(clusterName string, config cluster.TrackerConfig) error {
	cluster, ok := c.Cluster(clusterName)
	if !ok {
		return unknownClusterErr(clusterName)
	}
	cluster.UpdateTracker(config)
	return nil
}

// Upload file to the cluster group with specified return filename extension.
// The uploaded file cannot be appended. If you want to append bytes afterwards,
// use method UploadAppender.
//
// Upload is a wrapper of DefaultClient.Upload
func Upload(clusterName, group, ext string, b []byte) (string, error) {
	return DefaultClient.Upload(clusterName, group, ext, b)
}

// Upload file to the cluster group with specified return filename extension.
// The uploaded file cannot be appended. If you want to append bytes afterwards,
// use method UploadAppender.
func (c *Client) Upload(clusterName, group, ext string, b []byte) (string, error) {
	cluster, ok := c.Cluster(clusterName)
	if !ok {
		return "", unknownClusterErr(clusterName)
	}
	return cluster.Upload(b, group, ext)
}

// UploadAppender upload a file which can be appended bytes to.
//
// UploadAppender is a wrapper of DefaultClient.UploadAppender.
func UploadAppender(clusterName, group, ext string, b []byte) (string, error) {
	return DefaultClient.UploadAppender(clusterName, group, ext, b)
}

// UploadAppender upload a file which can be appended bytes to.
func (c *Client) UploadAppender(clusterName, group, ext string, b []byte) (string, error) {
	cluster, ok := c.Cluster(clusterName)
	if !ok {
		return "", unknownClusterErr(clusterName)
	}
	return cluster.UploadAppender(b, group, ext)
}

// UploadSlave upload as a slave of master file. Returned filename is with format {master}{suffix}.{ext}.
//
// UploadSlave is a wrapper of DefaultClient.UploadSlave.
func UploadSlave(clusterName, master, suffix, ext string, b []byte) (string, error) {
	return DefaultClient.UploadSlave(clusterName, master, suffix, ext, b)
}

// UploadSlave upload as a slave of master file. Returned filename is with format {master}{suffix}.{ext}
func (c *Client) UploadSlave(clusterName, master, suffix, ext string, b []byte) (string, error) {
	cluster, ok := c.Cluster(clusterName)
	if !ok {
		return "", unknownClusterErr(clusterName)
	}
	return cluster.UploadSlave(b, master, suffix, ext)
}
