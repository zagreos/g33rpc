package xclient

import (
	"log"
	"net/http"
	"strings"
	"time"
)

type G33RegistryDiscovery struct {
	*MultiServersDiscovery
	registry		string	// the address of the registry
	timeout        	time.Duration
	lastUpdate 		time.Time
}

const defaultUpdateTimeout = time.Second * 10

func NewG33RegistryDiscovery(registerAddr string, timeout time.Duration) *G33RegistryDiscovery{
	if timeout == 0 {
		timeout = defaultUpdateTimeout
	}
	d := &G33RegistryDiscovery{
		MultiServersDiscovery: 	NewMultiServerDiscovery(make([]string, 0)),
		registry:				registerAddr,
		timeout:				timeout,
	}
	return d
}

func (d *G33RegistryDiscovery) Update(servers []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.servers = servers
	d.lastUpdate = time.Now()
	return nil
}

func (d *G33RegistryDiscovery) Refresh() error {
	d.mu.Lock()
	defer d.mu.Lock()
	if d.lastUpdate.Add(d.timeout).After(time.Now()) {
		return nil
	}
	log.Println("rpc registry : refresh servers from registry", d.registry)
	resp, err := http.Get(d.registry)
	if err != nil {
		log.Println("rpc registry refresh err : ", err)
		return err
	}
	servers := strings.Split(resp.Header.Get("X-G33rpc-Servers"), ",")
	d.servers = make([]string, len(servers))
	for _, server := range servers {
		if strings.TrimSpace(server) != "" {
			d.servers = append(d.servers, strings.TrimSpace(server))
		}
	}
	d.lastUpdate = time.Now()
	return nil
}

func (d *G33RegistryDiscovery) Get(mode SelectMode) (string ,error) {
	if err := d.Refresh(); err != nil {
		return "", err
	}
	return d.MultiServersDiscovery.Get(mode)
}

func (d *G33RegistryDiscovery) GetAll() ([]string, error) {
	if err := d.Refresh(); err != nil {
		return nil, err
	}
	return d.MultiServersDiscovery.GetAll()
}

