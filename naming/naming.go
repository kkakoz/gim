package naming

import (
	"errors"
	"github.com/kkakoz/gim"
)

var (
	ErrNotFound = errors.New("service no found")
)

// Naming defined methods of the naming service
type Naming interface {
	Find(serviceName string, tags ...string) ([]gim.ServiceRegistration, error)
	Subscribe(serviceName string, callback func(services []gim.ServiceRegistration)) error
	Unsubscribe(serviceName string) error
	Register(service gim.ServiceRegistration) error
	Deregister(serviceID string) error
}
