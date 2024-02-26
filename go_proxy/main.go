package main

import (
	"crypto/tls"
	"flag"
	"log"

	"net/http"

	"github.com/muslimitsuhide/vk-security-hw/tree/main/go_proxy/proxy"
)

func main() {
	p := &proxy.Proxy{}

	certsDir := "certs"

	var protocol, crt, key string
	flag.StringVar(&crt, "crt", certsDir+"/ca.crt", "")
	flag.StringVar(&key, "key", certsDir+"/ca.key", "")
	flag.StringVar(&protocol, "protocol", "http", "")
	flag.Parse()

	server := &http.Server{
		Addr:         ":8080",
		Handler:      p,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	switch protocol {
	case "http":
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf(err.Error())
		}
		break
	case "https":
		if err := server.ListenAndServeTLS(crt, key); err != nil {
			log.Fatalf(err.Error())
		}
		break
	default:
		log.Println("http or https allowed")
		break
	}
}
