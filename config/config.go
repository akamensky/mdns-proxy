package config

import "flag"

func init() {
	listenAddr := flag.String("listenAddr", "0.0.0.0:80", "Specify the address to listen on")
	localSuffix := flag.String("localSuffix", "local", "Specify the local suffix used for mDNS network")

	flag.Parse()

	ListenAddr = *listenAddr
	LocalSuffix = *localSuffix
}

var (
	ListenAddr  string
	LocalSuffix string
)
