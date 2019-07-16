serverfinder-go 简单分布布服务器发现go语言实现库
===================================

### 项目说明：

在分布式应用中，怎样处理多个节点的互相感知问题，这是一门必修课。然而，应用有大有小，功课也可繁可简，本项目正是从简出发的一个工程实现。

很多项目在初创阶段，从简单性考虑，有时会采用配置文件来实现节点共享。然而，这却带来了一个很大的运维问题，因为每次节点变更，都需要手工变更配置文件，并进行节点间分发。更严重的是，节点异常掉线时，却无法感知。

本项目正是为了解决这个问题，但是又不至于太复杂。实现思路是利用redis做为代理人，所有节点定时向它上报自已的节点信息，并且定时从它那儿拉取所有节点信息。如果某个节点发生故障，redis会自动过期其数据，实现节点间动态感知分享。
```
graph LR
ServerA-->Redis
Redis-->ServerB
```
由于定时拉取对流量和CPU的资源消耗比较大，建议总节点数在500以内的中小型站点使用。

### 数据结构介绍：

###### 服务器节点：
Key的完整格式："SK_服务名_IP:PORT" <br>
Value的数据格式定义：<br>

```
type ServerNode struct {
	Name string		`json:"name"`
	Ip string		`json:"ip"`
	Port int		`json:"port"`
	Tags []string	`json:"tags"`
	LiveTick int64	`json:"livetick"`
}
```

字段说明：<br>
Name 服务器名字，用于节点类型的识别。<br>
Ip、Port 服务器工作的Ip、Port，用于外部向其连接。<br>
Tags 服务器的标签集合，可用于App识别、机房识别等用途。<br>
LiveTick 这个是服务器最新的写入时间，用于活跃性检查。<br>

###### 服务器节点推送者：

```
type PushWorker struct {
	server ServerNode

	expireSeconds int
	pushInterval int
	stopChan chan struct{}
	wg	sync.WaitGroup
}
```

字段说明：<br>
server 负责推送的服务器节点。<br>
expireSeconds 过期时间设置，默认30S。<br>
pushInterval 推送时间间隔设置，默认10S。<br>
stopChan、wg 协程同步。<br>

###### 所有服务器列表的容器索引：

```
type ServersContainer struct {
	servers []ServerNode
	lock sync.RWMutex

	// 根据名称建立索引
	mapNameIndexs map[string]*[]int

	// 根据标签建立索引
	mapTagIndexs map[string]*[]int
}

```

字段说明：<br>
servers 所有服务器元数据集合。<br>
lock 协程同步。<br>
mapNameIndexs 服务名索引。<br>
mapTagIndexs 服务标签索引。<br>
本结构不对外开放，防止被不小心修改。<br>

###### 服务器节点拉取者：

```
type PullWorker struct {
	servers []ServerNode

	pullInterval int
	stopChan chan struct{}
	wg	sync.WaitGroup

	container *ServersContainer
}
```

字段说明：<br>
servers 所有服务器元数据集合。<br>
pullInterval 定时拉取时间周期，默认10S。<br>
stopChan、wg 协程同步。<br>
container 服务器列表容器索引。<br>

### 方法介绍：

设置redis的地址：<br>
func SetRedisConfig(Ip string, Port int, Pass string, DB int)

创建服务器节点推送者：<br>
func NewPushWorker() (*PushWorker)

设置服务器节点名称：<br>
func (p *PushWorker) SetName(Name string)

设置服务器节点工作Ip:Port：<br>
func (p *PushWorker) SetIpPort(Ip string, Port int)

添加服务器节点Tag：<br>
func (p *PushWorker) AddTag(Tag string)

设置服务器节点过期时间：<br>
func (p *PushWorker) SetExpire(Seconds int)

设置服务器节点定时推送时间周期：<br>
func (p *PushWorker) SetPushInterval(Seconds int)

开启推送：<br>
func (p *PushWorker) StartPush()

停止推送：<br>
func (p *PushWorker) StopPush()

创建服务器列表拉取者：<br>
func NewPullWorker() (*PullWorker)

设置定时拉取时间周期：<br>
func (p *PullWorker) SetPullInterval(Seconds int)

启动拉取，并返回一个服务器列表容器指针：<br>
func (p *PullWorker) StartPull() (*ServersContainer)

停止拉取：<br>
func (p *PullWorker) StopPull()

获取所有服务器节点列表：<br>
func (c *ServersContainer) GetServers() ([]ServerNode)

获取指定过滤规则的服务器节点列表：<br>
func (c *ServersContainer) GetServersByFilters(Filters ...interface{}) ([]ServerNode)

过滤器：<br>
定义了以下两种过滤器：<br>

```
名称过滤器：
type NameFilter struct {
	Name string
}
标签过滤器：
type TagFilter struct {
	Tag string
}
```

可按需组合输出服务器元数据。

###### 最后举一个完整的例子：

```
package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"
    "serverfinder"
)

func main() {
    // 设置redis
    serverfinder.SetRedisConfig("127.0.0.1", 6379, "", 9)

    // 创建服务节点
    s1 := serverfinder.NewPushWorker()
    s1.SetName("gate") // 设置名字
    s1.SetIpPort("127.0.0.1", 1234)
    s1.AddTag("app_kdjie") // 打一个APP标签
    s1.AddTag("group_changsha") // 打一个机房标签
    s1.StartPush()

    // 通常只创建一个服务节点，这里创建两个是为了演示
    s2 := serverfinder.NewPushWorker()
    s2.SetName("cache") // 设置名字
    s2.SetIpPort("127.0.0.1", 5678)
    s2.AddTag("app_kdjie") // 打一个APP标签
    s2.AddTag("group_changsha") // 打一个机房标签
    s2.SetExpire(600) // 默认30S过期，这里设置600S
    s2.StartPush()

    // 创建拉取者
    p := serverfinder.NewPullWorker()
    // 启动拉取，返回容器索引
    c := p.StartPull()

    for {
        // 取所有服务器
        fmt.Println("all =>", c.GetServers())

        // 取名称为gate的服务器
        fmt.Println("gate =>", c.GetServersByFilters(serverfinder.NameFilter{"gate"}))

        // 取标签为"app_kdjie"的服务器
        fmt.Println("app_kdjie =>", c.GetServersByFilters(serverfinder.TagFilter{"app_kdjie"}))
        // 取标签为"group_changsha"的服务器
        fmt.Println("group_changsha =>", c.GetServersByFilters(serverfinder.TagFilter{"group_changsha"}))

        // 同时获取名称为gate，标签包含app_kdjie或group_changsha的服务器：
        fmt.Println("gate,app_kdjie,group_changsha=>",
            c.GetServersByFilters(serverfinder.NameFilter{"gate"}, serverfinder.TagFilter{"app_kdjie"}, serverfinder.TagFilter{"group_changsha"}))

        time.Sleep(time.Second*10)
    }

    s := make(chan os.Signal, 1)
    signal.Notify(s, os.Interrupt, syscall.SIGTERM)
    <-s

    s1.StopPush()
    s2.StopPush()
    p.StopPull()

    fmt.Println("ok")
}
```
