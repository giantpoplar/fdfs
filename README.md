# fdfs

[![GoDoc](https://godoc.org/github.com/giantpoplar/fdfs?status.png)](https://godoc.org/github.com/giantpoplar/fdfs)
![Travis (.org) branch](https://img.shields.io/travis/giantpoplar/fdfs/master.svg)

FDFS is a pure golang client for [FastDFS](https://github.com/happyfish100/fastdfs).

### Main Features

- 支持多个相互独立的FDFS集群
- 使用可精确控制连接数的连接池管理Tracker和Storage连接
- 独立设置Tracker和Storage的配置，支持配置热更新

### Getting Started

#### 仅有1个FastDFS集群

假设集群Tracker地址为127.0.0.1:22122，仅需简单一步就可完成客户端初始化：
```
err := cluster.Init([]string{"127.0.0.1:22122"}, cluster.TrackerConfig{}, cluster.StorageConfig{})
```
上述初始化将使用默认配置，你也可以根据需要自定义配置:
```
myTrackerConfig := cluster.TrackerConfig{
    PoolConfig: pool.Config{
        CacheMethod: pool.FIFO,        // 连接池管理方式，先进连接将先被使用
        InitCap:     1,                // 初始创建的连接数
        MaxCap:      3,                // 最大可创建连接数
        IdleTimeout: 30 * time.Second, // 连接空闲时间，超过此时间连接将被关闭
        WaitTimeout: 3 * time.Second,  // 等待空闲连接超时时间
        IOTimeout:   30 * time.Second, // 读写超时时间
        DialTimeout: 30 * time.Second, // 拨号超时时间
    },
}
myStorageConfig := cluster.StorageConfig{
    // 最大可下载文件大小，当请求下载文件大于该限制时直接失败.
    // 设置该值是为了避免请求下载超大文件对客户端和服务端冲击
    DownloadSizeLimit: 128 * 1024 * 1024,
    PoolConfig: pool.Config{
        CacheMethod: pool.FILO,        // 连接池管理方式，先进连接将后被使用
        InitCap:     2,                // 初始创建的连接数
        MaxCap:      8,                // 最大可创建连接数
        IdleTimeout: 30 * time.Second, // 连接空闲时间，超过此时间连接将被关闭
        WaitTimeout: 3 * time.Second,  // 等待空闲连接超时时间
        IOTimeout:   30 * time.Second, // 读写超时时间
        DialTimeout: 30 * time.Second, // 拨号超时时间
    },
}
err := cluster.Init([]string{"127.0.0.1:22122"}, myTrackerConfig, myStorageConfig)
```
你也可以根据系统运行状态实时热更新配置:
```
// 更新所有Tracker配置
cluster.UpdateTracker(your_tracker_config)
// 更新所有属于g1的storage配置
cluster.UpdateStorageGroup("g1", your_storage_config)
```
初始化完成后，就可以执行上传、下载、删除等操作:
```
// 上传文件到storage组g1,并指定返回文件后缀为jpg
fid, err := cluster.Upload("g1", "jpg", b)
// 下载文件
b, err := cluster.Download("g1/M01/DE/79/CgIG6VuXIoeAbiwbAAAIIRe5FG4412.jpg")
// 删除文件
err := cluster.Delete("g1/M01/DE/79/CgIG6VuXIoeAbiwbAAAIIRe5FG4412.jpg")
```

#### 有多个独立的FastDFS集群

如果你有两个独立的FastDFS集群, 命名为fdfs_cluster1和fdfs_cluster2，集群1的tracker地址为:10.15.25.46:22122和10.15.25.47:22122，集群2的tracker地址为:10.28.89.12:22122和10.28.89.13:22122，可按如下步骤初始化:
```
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
```
此时，上传、下载、删除等操作变为:
```
// 上传文件到fdfs_cluster1 storage组g1,并指定返回文件后缀为jpg
fid, err := fdfs.Upload("fdfs_cluster1", "g1", "jpg", b)
// 从fdfs_cluster1下载文件
b, err := fdfs.Download("fdfs_cluster1", "g1/M01/DE/79/CgIG6VuXIoeAbiwbAAAIIRe5FG4412.jpg")
// 删除fdfs_custer1中的文件
err := fdfs.Delete("fdfs_cluster1", "g1/M01/DE/79/CgIG6VuXIoeAbiwbAAAIIRe5FG4412.jpg")
// 更新fdfs_cluster1的tracker配置
fdfs.UpdateTracker("fdfs_cluter1", your_tracker_config)
// 更新fdfs_cluster1属于g1的所有storage配置
fdfs.UpdateStorageGroup("fdfs_cluster1", "g1", your_storage_config)
```
### API

[GoDoc](https://www.godoc.org/github.com/giantpoplar/fdfs)