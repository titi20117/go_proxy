package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
)

var severCount = 0

const (
	SERVER = "https://habr.com/ru/"
	PORT   = "1000"
)

type transport struct {
	http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	resp, err = t.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	for _, el := range bytes.Split(b, []byte(" ,.")) {
		if len(string(el)) == 6 {
			b = bytes.Replace(b, []byte(string(el)), []byte(string(el)+"\u2122"), -1)
		}

	}
	body := ioutil.NopCloser(bytes.NewReader(b))
	resp.Body = body
	resp.ContentLength = int64(len(b))
	resp.Header.Set("Content-Length", strconv.Itoa(len(b)))
	return resp, nil
}

var _ http.RoundTripper = &transport{}

func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	url, _ := url.Parse(target)

	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.Transport = &transport{http.DefaultTransport}

	proxy.ServeHTTP(res, req)
}

func logRequestPayload(proxyURL string) {
	log.Printf("proxy_url: %s\n", proxyURL)
}

func getProxyURL() string {
	var servers = []string{SERVER}
	server := servers[severCount]
	severCount++

	if severCount >= len(servers) {
		severCount = 0
	}
	return server
}

func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	url := getProxyURL()
	logRequestPayload(url)
	serveReverseProxy(url, res, req)
}

func main() {
	http.HandleFunc("/", handleRequestAndRedirect)
	log.Fatal(http.ListenAndServe(":"+PORT, nil))
}
