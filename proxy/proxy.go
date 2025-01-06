package proxy

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"mdns-proxy/config"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var (
	proxies = make(map[string]*httputil.ReverseProxy)
)

func init() {
	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

		handler := func() http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				proxy, ok := proxies[r.Host]
				if !ok {
					http.Error(w, "Not found", http.StatusNotFound)
					return
				}

				r.Header.Set("X-Forwarded-Host", r.Host)
				r.Header.Set("X-Forwarded-Proto", "http")
				r.Header.Set("X-Real-IP", strings.SplitN(r.RemoteAddr, ":", 2)[0])

				proxy.ServeHTTP(w, r)
			})
		}

		srv := &http.Server{
			Addr:    config.ListenAddr,
			Handler: handler(),
		}

		go func() {
			<-signalChan
			slog.Info("Received signal, shutting down server...")
			if err := srv.Shutdown(context.Background()); err != nil {
				slog.Error("Error during server shutdown", slog.String("error", err.Error()))
			}
		}()

		if err := srv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				os.Exit(0)
			} else {
				slog.Error("Server error", slog.String("error", err.Error()))
				os.Exit(1)
			}
		}
	}()
}

func SetRules(rules map[string]string) error {
	var wrappedError error

	newProxies := make(map[string]*httputil.ReverseProxy)

	for host, addr := range rules {
		parsedURL, err := url.Parse(addr)
		if err != nil {
			if wrappedError == nil {
				wrappedError = fmt.Errorf("failed to parse URL %s: %w", addr, err)
			} else {
				wrappedError = fmt.Errorf("%w; failed to parse URL %s: %w", wrappedError, addr, err)
			}
			continue
		}
		newProxies[host] = httputil.NewSingleHostReverseProxy(parsedURL)
		newProxies[host].ErrorLog = log.New(io.Discard, "", 0)
		newProxies[host].ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {}
		newProxies[host].Transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	proxies = newProxies

	return wrappedError
}
