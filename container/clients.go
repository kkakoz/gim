package container

import (
	"github.com/kkakoz/gim"
	"github.com/kkakoz/gim/pkg/mapx"
)

type IClientMap interface {
	Add(client gim.Client)
	Remove(id string)
	Get(id string) (gim.Client, bool)
	Services(kvs []string) []gim.Service
}

type clientMap struct {
	m *mapx.SyncMap[string, gim.Client]
}

func (c *clientMap) Add(client gim.Client) {
	c.m.Add(client.ServiceID(), client)
}

func (c *clientMap) Remove(id string) {
	c.m.Delete(id)
}

func (c *clientMap) Get(id string) (gim.Client, bool) {
	return c.m.Get(id)
}

func (c *clientMap) Services(kvs []string) []gim.Service {
	res := []gim.Service{}
	for _, v := range c.m.Values() {
		if v.GetMeta()[kvs[0]] == kvs[1] {
			res = append(res, v)
		}
	}
	return res
}

func NewClients() *clientMap {
	return &clientMap{m: mapx.NewSyncMap[string, gim.Client]()}
}
