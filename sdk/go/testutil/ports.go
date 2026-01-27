package testutil

import (
	"net"
	"sync"
)

// GetUsablePorts asks the kernel for unused TCP ports
func GetUsablePorts(count int) (*UsablePorts, error) {
	var ports []int
	listeners := make([]*net.TCPListener, 0, count)

	defer func() {
		for _, l := range listeners {
			_ = l.Close()
		}
	}()

	for i := 0; i < count; i++ {
		addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
		if err != nil {
			return nil, err
		}

		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return nil, err
		}

		listeners = append(listeners, l)
		ports = append(ports, l.Addr().(*net.TCPAddr).Port)
	}

	return usablePorts(ports), nil
}

// UsablePorts yields a finite sequence of ephemeral TCP port numbers discovered at creation time.
// Safe for concurrent use; do not copy this struct after first use.
type UsablePorts struct {
	lock  sync.Mutex
	idx   int
	ports []int
}

func usablePorts(ports []int) *UsablePorts {
	return &UsablePorts{
		idx:   0,
		ports: ports,
	}
}

func (p *UsablePorts) MustGetPort() int {
	defer p.lock.Unlock()
	p.lock.Lock()

	if p.idx == len(p.ports) {
		panic("no ports available")
	}

	port := p.ports[p.idx]
	p.idx++

	return port
}
