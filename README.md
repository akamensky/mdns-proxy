# mDNS-Proxy service

## Overview

A service that discovers applications running on Docker, then publishes their name as another hostname on mDNS and sets up a reverse proxy. 
This is a simple enough solution for many home labs that allow forgetting about different ports.

It currently has some limitations:

- Since mDNS is a multicast, if it is running in the container then it should be using the host network (any other solution is overcomplicating it IMO).
- Since it discovers applications on Docker, it needs docker Unix socket accessible to it, so it needs to be either mounted in container, or be running as root as as user who is a member of docker group.
- For a quick solution it uses `github.com/pion/mdns/v2`, which _currently_ does not allow adding/removing published names without restarting the server, this means that every time new name added (or one removed) the mDNS server will be restarted, which may cause brief interruptions (contributions to change this are welcome)

## Install

### With Docker

In your compose file add a service as such:

```yaml
  mdns-proxy:
    image: "ghcr.io/akamensky/mdns-proxy:latest"
    container_name: "mdns-proxy"
    network_mode: "host"
    ports:
      - "80:80"
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock:ro"
```

### Without Docker

You will need Go 1.23 or newer

1. `git clone https://github.com/akamensky/mdns-proxy.git`
2. `cd mdns-proxy`
3. `go build .`
4. Use produced binary

## Usage

```
$ ./mdns-proxy -h
  -listenAddr string
    	Specify the address to listen on (default "0.0.0.0:80")
```

## License

See LICENSE file.
