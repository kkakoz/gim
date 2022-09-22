package gim

import "github.com/kkakoz/gim/pkg/mapx"

// IChannelMap 连接管理器，Server在内部会自动管理连接的生命周期
type IChannelMap interface {
	Add(channel IChannel)
	Remove(id string)
	Get(id string) (IChannel, bool)
	All() []IChannel
}

type channelMap struct {
	channels *mapx.SyncMap[string, IChannel]
}

func NewChannelMap() *channelMap {
	return &channelMap{channels: mapx.NewSyncMap[string, IChannel]()}
}

func (c *channelMap) Add(channel IChannel) {
	c.channels.Add(channel.ID(), channel)
}

func (c *channelMap) Remove(id string) {
	c.channels.Delete(id)
}

func (c *channelMap) Get(id string) (IChannel, bool) {
	return c.channels.Get(id)
}

func (c *channelMap) All() []IChannel {
	return c.channels.Values()
}

func NewChannels() *channelMap {
	return &channelMap{}
}
