package main

import (
	"log/slog"
	"mdns-proxy/docker"
	"mdns-proxy/mdns"
	"mdns-proxy/proxy"
	"mdns-proxy/service"
	"sort"
	"time"
)

var currentServices []service.Service

func main() {
	do()
	t := time.NewTicker(5 * time.Second)
	for range t.C {
		do()
	}
}

func do() {
	newServices, err := docker.Discover()
	if err != nil {
		slog.Error(err.Error())
		return
	}
	sort.Slice(newServices, func(i, j int) bool {
		return newServices[i].Name() < newServices[j].Name()
	})

	// compare service slices and if same skip as it is a noop
	if serviceSlicesEqual(newServices, currentServices) {
		return
	}

	currentServices = newServices

	newNames := make([]string, 0, len(newServices))
	newProxyRules := make(map[string]string)
	// Generate new names list by appending `.local` to name
	for _, svc := range newServices {
		name := svc.Name() + ".local"
		newNames = append(newNames, name)
		newProxyRules[name] = svc.Address()
	}

	// Set new names in mDNS server
	err = mdns.Set(newNames)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	// Setup proxy rules
	if err := proxy.SetRules(newProxyRules); err != nil {
		slog.Error(err.Error())
	}
}

func serviceSlicesEqual(a, b []service.Service) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Name() != b[i].Name() {
			return false
		}
		if a[i].Address() != b[i].Address() {
			return false
		}
	}
	return true
}
