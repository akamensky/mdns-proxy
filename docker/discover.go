package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"log/slog"
	"mdns-proxy/service"
	"net/url"
	"regexp"
	"strings"
)

var nameRe = regexp.MustCompile(`^[a-z0-9]+$`)

type dockerService struct {
	name string
	addr string
}

func (s *dockerService) Name() string {
	return s.name
}

func (s *dockerService) Address() string {
	return s.addr
}

func Discover() ([]service.Service, error) {
	// Create a Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	// List all running containers
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{
		All:    true,
		Latest: true,
	})
	if err != nil {
		return nil, err
	}

	var result []service.Service

	for _, c := range containers {
		// Skip containers that are not running
		if c.State != "running" {
			continue
		}

		s := &dockerService{}

		// Filter based on labels
		// the labels we are looking for are:
		// - mdns-proxy.enable=true to enable proxying and mdns publishing
		// - mdns-proxy.name=serviceName that sets service name, then published on mDNS as serviceName.local
		// - mdns-proxy.address=http://... that is used as address to proxy requests to
		if val, ok := c.Labels["mdns-proxy.enable"]; !ok || !strings.EqualFold(val, "true") {
			// proxying is not enabled, so skipping
			continue
		}
		if val, ok := c.Labels["mdns-proxy.name"]; !ok {
			slog.Error(fmt.Sprintf("Container %s has no mdns-proxy.name label", c.Names[0]))
			continue
		} else if !nameRe.MatchString(val) {
			slog.Error(fmt.Sprintf("Container %s has invalid mdns-proxy.name label: %s", c.Names[0], val))
			continue
		} else {
			s.name = val
		}

		// Check for duplicate service names in the resulting list
		for _, existing := range result {
			if existing.Name() == s.name {
				slog.Error(fmt.Sprintf("Duplicate service name detected: %s", s.name))
				continue
			}
		}

		if val, ok := c.Labels["mdns-proxy.address"]; !ok {
			slog.Error(fmt.Sprintf("Container %s has no mdns-proxy.address label", c.Names[0]))
			continue
		} else if _, err := url.Parse(val); err != nil {
			slog.Error(fmt.Sprintf("Container %s has invalid mdns-proxy.address label: %s", c.Names[0], val))
			continue
		} else {
			s.addr = val
		}

		result = append(result, s)
	}

	return result, nil
}
