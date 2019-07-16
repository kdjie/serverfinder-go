package serverfinder

import (
	"fmt"
	"time"
	"sync"
	"encoding/json"
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
	val, err := pCmd.Result()
	if err != nil {
		return err
	}

	if len(val) > 0 {
		// 再取values
		pCmd := client.MGet(val...)
		vals, err := pCmd.Result()
		if err != nil {
			return err
		}

		p.servers = []ServerNode{}

		for _, val := range vals {
			if data, ok := val.(string); ok {
				server := ServerNode{}
				err = json.Unmarshal([]byte(data), &server)
				if err == nil {
					p.servers = append(p.servers, server)
				}
			}
		}

		//fmt.Println(p.servers)
		// 将结果放到索引容器中
		p.container.__update(p.servers)
	}

	return nil
}
