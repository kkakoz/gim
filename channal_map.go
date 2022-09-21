package gim

import "github.com/kkakoz/gim/pkg/mapx"

// IChannelMap 连接管理器，Server在内部会自动管理连接的生命周期
type IChannelMap interface {
	Add(channel Channel)
	Remove(id string)
	Get(id string) (Channel, bool)
	All() []Channel
}

type channelMap struct {
	channels *mapx.SyncMap[string, Channel]
}

func NewChannelMap() *channelMap {
	return &channelMap{channels: mapx.NewSyncMap[string, Channel]()}
}

func (c *channelMap) Add(channel Channel) {
	c.channels.Add(channel.ID(), channel)
}

func (c *channelMap) Remove(id string) {
	c.channels.Delete(id)
}

func (c *channelMap) Get(id string) (Channel, bool) {
	return c.channels.Get(id)
}

func (c *channelMap) All() []Channel {
	return c.channels.Values()
}

func NewChannels() *channelMap {
	return &channelMap{}
}
