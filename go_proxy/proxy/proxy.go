package proxy

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Proxy struct {
	db *sql.DB
}

func NewProxy(db *sql.DB) *Proxy {
	return &Proxy{db: db}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "CONNECT" {
		p.handleHTTPS(w, r)
	} else {
		p.handleHTTP(w, r)
	}
}

func (p *Proxy) handleHTTP(w http.ResponseWriter, r *http.Request) {
	r.RequestURI = ""
	r.Header.Del("Proxy-Connection")

	httpClient := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	proxyResponse, err := httpClient.Do(r)
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer proxyResponse.Body.Close()

	p.saveRequestResponse(r, proxyResponse)

	copyHeaders(w.Header(), proxyResponse.Header)
	w.WriteHeader(proxyResponse.StatusCode)
	io.Copy(w, proxyResponse.Body)
}

func copyHeaders(to, from http.Header) {
	for header, values := range from {
		for _, value := range values {
			to.Add(header, value)
		}
	}
}

func (p *Proxy) handleHTTPS(w http.ResponseWriter, r *http.Request) {
	connDest, err := p.connectHandshake(w, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	connSrc, _, err := hijacker.Hijack()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	go p.exchangeData(connDest, connSrc)
	go p.exchangeData(connSrc, connDest)
}

func (p *Proxy) saveRequestResponse(req *http.Request, resp *http.Response) {
	requestInfo := p.parseRequest(req)
	responseInfo := p.parseResponse(resp)

	tx, err := p.db.Begin()
	if err != nil {
		log.Println("Error beginning transaction:", err)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("INSERT INTO requests (request, response) VALUES (?, ?)", requestInfo, responseInfo)
	if err != nil {
		log.Println("Error inserting data into database:", err)
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Error committing transaction:", err)
	}
}

func (p *Proxy) parseRequest(r *http.Request) string {
	getParams := make(map[string]string)
	for key, values := range r.URL.Query() {
		getParams[key] = values[0]
	}

	headers := make(map[string]string)
	for key, values := range r.Header {
		headers[key] = values[0]
	}

	cookies := make(map[string]string)
	for _, cookie := range r.Cookies() {
		cookies[cookie.Name] = cookie.Value
	}

	postParams := make(map[string][]string)
	r.ParseForm()
	for key, values := range r.PostForm {
		postParams[key] = values
	}

	requestInfo := struct {
		Method     string              `json:"method"`
		Path       string              `json:"path"`
		GetParams  map[string]string   `json:"get_params"`
		Headers    map[string]string   `json:"headers"`
		Cookies    map[string]string   `json:"cookies"`
		PostParams map[string][]string `json:"post_params"`
	}{
		Method:     r.Method,
		Path:       r.URL.Path,
		GetParams:  getParams,
		Headers:    headers,
		Cookies:    cookies,
		PostParams: postParams,
	}

	jsonData, err := json.Marshal(requestInfo)
	if err != nil {
		log.Println("Error marshaling request data:", err)
		return ""
	}

	return string(jsonData)
}

func (p *Proxy) parseResponse(resp *http.Response) string {
	headers := make(map[string]string)
	for key, values := range resp.Header {
		headers[key] = values[0]
	}

	responseInfo := struct {
		Code    int               `json:"code"`
		Message string            `json:"message"`
		Headers map[string]string `json:"headers"`
	}{
		Code:    resp.StatusCode,
		Message: resp.Status,
		Headers: headers,
	}

	jsonData, err := json.Marshal(responseInfo)
	if err != nil {
		log.Println("Error marshaling response data:", err)
		return ""
	}

	return string(jsonData)
}

func (p *Proxy) copyHeaders(to, from http.Header) {
	for header, values := range from {
		for _, value := range values {
			to.Add(header, value)
		}
	}
}

func (p *Proxy) exchangeData(to io.WriteCloser, from io.ReadCloser) {
	defer func() {
		to.Close()
		from.Close()
	}()

	io.Copy(to, from)
}

func (p *Proxy) connectHandshake(w http.ResponseWriter, r *http.Request) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", r.Host, 10000*time.Millisecond)
	if err != nil {
		return nil, err
	}

	w.WriteHeader(http.StatusOK)
	return conn, nil
}
