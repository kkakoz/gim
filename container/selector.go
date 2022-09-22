package container

import (
	"github.com/kkakoz/gim"
	"github.com/kkakoz/gim/proto/pkt"
	"hash/crc32"
)

type Selector interface {
	Lookup(*pkt.Header, []gim.Service) string
}

type HashSelector struct {
}

func (h *HashSelector) Lookup(header *pkt.Header, services []gim.Service) string {
	ll := len(services)
	code := HashCode(header.ChannelId)
	return services[code%ll].ServiceID()
}

func HashCode(key string) int {
	hash32 := crc32.NewIEEE()
	hash32.Write([]byte(key))
	return int(hash32.Sum32())
}
