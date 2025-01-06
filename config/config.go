package config

import "flag"

func init() {
	listenAddr := flag.String("listenAddr", "0.0.0.0:80", "Specify the address to listen on")

	flag.Parse()

	ListenAddr = *listenAddr
}

var (
	ListenAddr string
)
