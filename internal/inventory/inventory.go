package inventory

import (
	"os"

	"github.com/Higangssh/homebutler/internal/config"
	"github.com/Higangssh/homebutler/internal/docker"
	"github.com/Higangssh/homebutler/internal/ports"
	"github.com/Higangssh/homebutler/internal/system"
)

// Inventory holds the collected topology for a single server.
type Inventory struct {
	ServerName string             `json:"server_name"`
	Host       string             `json:"host"`
	System     *system.StatusInfo `json:"system"`
	Containers []docker.Container `json:"containers"`
	Ports      []ports.PortInfo   `json:"ports"`
	Warnings   []string           `json:"warnings,omitempty"`
}

// CollectFuncs allows injecting data sources for testing.
type CollectFuncs struct {
	StatusFn     func() (*system.StatusInfo, error)
	DockerListFn func() ([]docker.Container, error)
	PortsListFn  func() (*ports.Result, error)
}

// DefaultCollectFuncs returns the real system/docker/ports functions.
func DefaultCollectFuncs() CollectFuncs {
	return CollectFuncs{
		StatusFn:     system.Status,
		DockerListFn: docker.List,
		PortsListFn:  ports.List,
	}
}

// Collect gathers inventory for the local server.
// Docker and ports failures are recorded as warnings, not errors.
func Collect(cfg *config.Config, fns CollectFuncs) (*Inventory, error) {
	inv := &Inventory{
		Containers: []docker.Container{},
		Ports:      []ports.PortInfo{},
	}

	// Determine server name and host from config.
	inv.ServerName, inv.Host = resolveServer(cfg)

	// System status is required.
	info, err := fns.StatusFn()
	if err != nil {
		return nil, err
	}
	inv.System = info

	// Docker: best-effort.
	containers, err := fns.DockerListFn()
	if err != nil {
		inv.Warnings = append(inv.Warnings, "docker: "+err.Error())
	} else {
		inv.Containers = containers
	}

	// Ports: best-effort.
	result, err := fns.PortsListFn()
	if err != nil {
		inv.Warnings = append(inv.Warnings, "ports: "+err.Error())
	} else {
		inv.Ports = result.Ports
	}

	return inv, nil
}

// resolveServer picks the local server name/host from config,
// falling back to os.Hostname.
func resolveServer(cfg *config.Config) (name, host string) {
	if cfg != nil {
		for _, s := range cfg.Servers {
			if s.Local {
				return s.Name, s.Host
			}
		}
	}
	h, _ := os.Hostname()
	return h, h
}
