package serverfinder

import (
	"sync"
)

// 所有服务器的容器索引
type ServersContainer struct {
	servers []ServerNode
	lock sync.RWMutex

	// 根据名称建立索引
	mapNameIndexs map[string]*[]int

	// 根据标签建立索引
	mapTagIndexs map[string]*[]int
}

// 创建容器索引对象
func NewServersContainer() (*ServersContainer) {
	return &ServersContainer{
		servers : []ServerNode{},
		mapNameIndexs : make(map[string]*[]int),
		mapTagIndexs : make(map[string]*[]int),
	}
}

// 更新容器索引
func (c *ServersContainer) __update(Servers []ServerNode) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// 深度拷贝
	c.servers = []ServerNode{}
	c.servers = append(c.servers, Servers...)

	// 创建索引
	c.__createNameIndexs()
	c.__createTagIndexs()

/*
	fmt.Println("servers", c.servers)
	for k,v := range c.mapNameIndexs {
		fmt.Println("\t", k, *v)
	}
	for k,v := range c.mapTagIndexs {
		fmt.Println("\t", k, *v)
	}
*/
}

func (c *ServersContainer) __createNameIndexs() {
	c.mapNameIndexs = make(map[string]*[]int)

	for index, server := range c.servers {
		_, ok := c.mapNameIndexs[server.Name]
		if !ok {
			c.mapNameIndexs[server.Name] = &([]int{})
		}

		pIndexs, _ := c.mapNameIndexs[server.Name]
		*pIndexs = append(*pIndexs, index)
	}
}

func (c *ServersContainer) __createTagIndexs() {
	c.mapTagIndexs = make(map[string]*[]int)

	for index, server := range c.servers {
		for _, Tag := range server.Tags {
			_, ok := c.mapTagIndexs[Tag]
			if !ok {
				c.mapTagIndexs[Tag] = &([]int{})
			}

			pIndexs, _ := c.mapTagIndexs[Tag]
			*pIndexs = append(*pIndexs, index)
		}
	}
}

func (c *ServersContainer) __getServersByIndexs(Indexs []int) ([]ServerNode) {
	Servers := []ServerNode{}

	for _, Index := range Indexs {
		if Index >= 0 && Index < len(c.servers) {
			Servers = append(Servers, c.servers[Index])
		}
	}

	return Servers
}

// 获取所有服务器元数据
func (c *ServersContainer) GetServers() ([]ServerNode) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	Servers := []ServerNode{}
	Servers = append(Servers, c.servers...)
	return Servers
}

func __checkRepeat(Indexs []int, Idx int) (bool) {
	for _, Index := range Indexs {
		if Index == Idx {
			return true
		}
	}
	return false
}

// 名称过滤器
type NameFilter struct {
	Name string
}

// 标签过滤器
type TagFilter struct {
	Tag string
}

// 获取指定过滤规则的服务器元数据
func (c *ServersContainer) GetServersByFilters(Filters ...interface{}) ([]ServerNode) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	// 对名称、标签过滤器进行分类汇总，
	// 分别统计出各自的过滤器数量和索引集合

	nameIndexs := []int{}
	tagIndexs := []int{}
	nameOperate := false
	tagOperate := false

	for _, Filter := range Filters {
		switch Filter.(type) {
		case NameFilter:
			nameOperate = true
			pNF := Filter.(NameFilter)
			if pIndexs, ok := c.mapNameIndexs[pNF.Name]; ok {
				nameIndexs = append(nameIndexs, (*pIndexs)...)
			}
		case TagFilter:
			tagOperate = true
			pTF := Filter.(TagFilter)
			if pIndexs, ok := c.mapTagIndexs[pTF.Tag]; ok {
				for _, Index := range *pIndexs {
					// 去重
					if !__checkRepeat(tagIndexs, Index) {
						tagIndexs = append(tagIndexs, Index)
					}
				}
			}
		}
	}

	// 然后根据过滤器分类，进行按类别取交集

	resultIndexs := []int{}
	andOperate := false

	if nameOperate {
		if !andOperate {
			resultIndexs = append(resultIndexs, nameIndexs...)
		} else {
			joinIndexs := []int{}
			for _, IndexA := range resultIndexs {
				for _, IndexB := range nameIndexs {
					if IndexA == IndexB {
						joinIndexs = append(joinIndexs, IndexA)
						continue
					}
				}
			}
			resultIndexs = []int{}
			resultIndexs = append(resultIndexs, joinIndexs...)
		}

		andOperate = true
	}

	if tagOperate {
		if !andOperate {
			resultIndexs = append(resultIndexs, tagIndexs...)
		} else {
			joinIndexs := []int{}
			for _, IndexA := range resultIndexs {
				for _, IndexB := range tagIndexs {
					if IndexA == IndexB {
						joinIndexs = append(joinIndexs, IndexA)
						continue
					}
				}
			}
			resultIndexs = []int{}
			resultIndexs = append(resultIndexs, joinIndexs...)
		}

		andOperate = true
	}

	// 根据索引取服务器
	return c.__getServersByIndexs(resultIndexs)
}
