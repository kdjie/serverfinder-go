package serverfinder

import (
	"fmt"
	"time"
	"sync"
	"encoding/json"
)

// 推送者结构定义
type PushWorker struct {
	server ServerNode

	expireSeconds int
	pushInterval int
	stopChan chan struct{}
	wg	sync.WaitGroup
}

// 创建一个推送者
func NewPushWorker() (*PushWorker) {
	return &PushWorker{
		server : ServerNode{
			Name :"",
			Ip : "",
			Port : 0,
			Tags : []string{},
			LiveTick : 0,
		},
		expireSeconds : 30,
		pushInterval : 10,
		stopChan : make(chan struct{}),
		wg : sync.WaitGroup{},
	}
}

// 设置服务器名称
func (p *PushWorker) SetName(Name string) {
	p.server.Name = Name
}

// 设置服务器工作的Ip:Port
func (p *PushWorker) SetIpPort(Ip string, Port int) {
	p.server.Ip = Ip
	p.server.Port = Port
}

// 添加服务器的Tag
func (p *PushWorker) AddTag(Tag string) {
	p.server.Tags = append(p.server.Tags, Tag)
}

// 设置服务器过期时间
func (p *PushWorker) SetExpire(Seconds int) {
	p.expireSeconds = Seconds
}

// 设置定时推送时间周期
func (p *PushWorker) SetPushInterval(Seconds int) {
	p.pushInterval = Seconds
}

// 开启推送
func (p *PushWorker) StartPush() {
	p.wg.Add(1)
	go p.__loop()
}

// 停止推送
func (p *PushWorker) StopPush() {
	close(p.stopChan)
	p.wg.Wait()
}

// 推送者工作协程
func (p *PushWorker) __loop() {
	defer p.wg.Done()
	defer p.__delete()

	p.__push()

	for {
		select {
		case <- p.stopChan:
			return
		case <- time.After( time.Second * time.Duration(p.pushInterval) ):
			p.__push()
		}
	}
}

// 格式化服务器节点的key
func (p *PushWorker) __toNodeKey() (string) {
	return fmt.Sprintf("%s_%s_%s:%d", SERVER_KEY_PREFIX, p.server.Name, p.server.Ip, p.server.Port)
}

// 格式化服务器节点为json
func (p *PushWorker) __toNodeJson() (string) {
	p.server.LiveTick = time.Now().Unix()

	data, err := json.Marshal(p.server)
	if err != nil {
		return ""
	}

	return string(data)
}

// 执行一次推送
func (p *PushWorker) __push() (error) {
	client := __getRedicClient()
	defer client.Close()

	pCmd := client.Set(p.__toNodeKey(), p.__toNodeJson(), time.Second * time.Duration(p.expireSeconds))
	_, err := pCmd.Result()
	if err != nil {
		return err
	}

	return nil
}

// 执行节点删除
func (p *PushWorker) __delete() {
	client := __getRedicClient()
	defer client.Close()

	client.Del( p.__toNodeKey() )
}
