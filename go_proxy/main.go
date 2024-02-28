package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"log"
	"net/http"

	"github.com/muslimitsuhide/vk-security-hw/proxy"
)

func main() {
	certsDir := "certs"

	var protocol, crt, key string
	flag.StringVar(&crt, "crt", certsDir+"/ca.crt", "")
	flag.StringVar(&key, "key", certsDir+"/ca.key", "")
	flag.StringVar(&protocol, "protocol", "http", "")
	flag.Parse()

	db, err := sql.Open("sqlite3", "./requests.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	server := &http.Server{
		Addr:         ":8080",
		Handler:      proxy.NewProxy(db),
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	switch protocol {
	case "http":
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf(err.Error())
		}
	case "https":
		if err := server.ListenAndServeTLS(crt, key); err != nil {
			log.Fatalf(err.Error())
		}
	default:
		log.Println("http or https allowed")
	}
}
