package mdns

import (
	"fmt"
	"github.com/pion/mdns/v2"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"log/slog"
	"net"
	"sync"
)

var (
	srv   *mdns.Conn
	names = make([]string, 0)
	lock  = new(sync.Mutex)
)

func init() {
	s, err := start(names)
	if err != nil {
		panic(err)
	}
	srv = s
}

func Names() []string {
	lock.Lock()
	defer lock.Unlock()

	newNames := make([]string, 0, len(names))
	for _, name := range names {
		newNames = append(newNames, name)
	}
	return newNames
}

func Set(nameset []string) error {
	lock.Lock()
	defer lock.Unlock()

	s, err := start(nameset)
	if err != nil {
		return err
	}

	names = nameset
	srv = s

	slog.Info(fmt.Sprintf("Set %v", nameset))

	return nil
}

func Add(name string) error {
	lock.Lock()
	defer lock.Unlock()

	// if name already exists on the list, then skip as this is a noop
	newNames := make([]string, 0, len(names)+1)
	for _, existingName := range names {
		if existingName == name {
			return nil
		}
		newNames = append(newNames, existingName)
	}
	newNames = append(newNames, name)

	s, err := start(newNames)
	if err != nil {
		return err
	}

	names = newNames
	srv = s

	slog.Info(fmt.Sprintf("Added %s", name))

	return nil
}

func Remove(name string) error {
	lock.Lock()
	defer lock.Unlock()

	// Check if the name exists in the list
	found := false
	newNames := make([]string, 0, len(names))
	for _, existingName := range names {
		if existingName == name {
			found = true
		} else {
			newNames = append(newNames, existingName)
		}
	}

	if !found {
		// if the name doesn't exist in the list, it's a noop
		return nil
	}

	s, err := start(newNames)
	if err != nil {
		return err
	}

	names = newNames
	srv = s

	slog.Info(fmt.Sprintf("Removed %s", name))

	return nil
}

func start(names []string) (*mdns.Conn, error) {
	// close the existing server
	if srv != nil {
		err := srv.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to close mDNS server: %w", err)
		}
	}

	var conn4 *ipv4.PacketConn
	addr4, err := net.ResolveUDPAddr("udp4", mdns.DefaultAddressIPv4)
	if err == nil {
		l4, err := net.ListenUDP("udp4", addr4)
		if err == nil {
			conn4 = ipv4.NewPacketConn(l4)
		}
	}

	var conn6 *ipv6.PacketConn
	addr6, err := net.ResolveUDPAddr("udp6", mdns.DefaultAddressIPv6)
	if err == nil {
		l6, err := net.ListenUDP("udp6", addr6)
		if err == nil {
			conn6 = ipv6.NewPacketConn(l6)
		}

	}

	// Start a new server with the updated list
	s, err := mdns.Server(conn4, conn6, &mdns.Config{
		LocalNames: names,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start mDNS server: %w", err)
	}
	return s, nil
}
