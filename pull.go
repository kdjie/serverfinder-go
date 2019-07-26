package serverfinder

import (
	"fmt"
	"time"
	"sync"
)

// 拉取者结构定义
type PullWorker struct {
	servers []ServerNode

	pullInterval int
	stopChan chan struct{}
	wg	sync.WaitGroup

	container *ServersContainer
}

// 创建拉取者对象
func NewPullWorker() (*PullWorker) {
	return &PullWorker{
		servers : []ServerNode{},
		pullInterval : 10,
		stopChan : make(chan struct{}),
		wg : sync.WaitGroup{},
		container : NewServersContainer(),
	}
}

// 设置定时拉取时间周期
func (p *PullWorker) SetPullInterval(Seconds int) {
	p.pullInterval = Seconds
}

// 启动拉取
func (p *PullWorker) StartPull() (*ServersContainer) {
	p.wg.Add(1)
	go p.__loop()

	return p.container
}

// 停止拉取
func (p *PullWorker) StopPull() {
	close(p.stopChan)
	p.wg.Wait()
}

// 拉取者工作协程
func (p *PullWorker) __loop() {
	defer p.wg.Done()

	p.__pull()

	for {
		select {
		case <- p.stopChan:
			return
		case <- time.After( time.Second * time.Duration(p.pullInterval) ):
			p.__pull()
		}
	}
}

// 执行一次远端服务器拉取
func (p *PullWorker) __pull() (error) {
	client := __getRedicClient()
	defer client.Close()

	// 先取keys
	pCmd := client.Keys( fmt.Sprintf("%s_*", SERVER_KEY_PREFIX) )
	keys, err := pCmd.Result()
	if err != nil {
		return err
	}

	// 为空时仍然更新容器
	if len(keys) == 0 {
		p.servers = []ServerNode{}
		p.container.__update(p.servers)
		return nil
	}

	// 再取values
	pCmd2 := client.MGet(keys...)
	vals, err := pCmd2.Result()
	if err != nil {
		return err
	}

	p.servers = []ServerNode{}

	for _, val := range vals {
		if data, ok := val.(string); ok {
			// 解析
			server := ServerNode{}
			if server.fromJson(data) {
				p.servers = append(p.servers, server)
			}
		}
	}

	//fmt.Println(p.servers)
	// 将结果放到索引容器中
	p.container.__update(p.servers)

	return nil
}
