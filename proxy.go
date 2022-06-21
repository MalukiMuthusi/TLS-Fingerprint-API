package main

import (
	"compress/zlib"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"
	"tlsapi/internal/session"

	"github.com/fatih/color"

	"bytes"
	"compress/gzip"
	"net/url"
	"strings"

	"github.com/Carcraftz/cclient"
	"github.com/andybalholm/brotli"

	http "github.com/Carcraftz/fhttp"

	tls "github.com/Carcraftz/utls"
)

//var client http.Client

func main() {

	tok := flag.String("token", "", "A token provided by the program provider")
	port := flag.String("port", "8082", "A port number (default 8082)")
	flag.Parse()

	// check the token exists in the database
	token, err := session.GetToken(*tok)
	if err != nil {
		panic("failed to check the provided token")
	}

	// check that the token is not expired

	expiry, err := time.Parse(time.RFC3339, token.ExpiryDate)
	if err != nil {
		panic("provide a valid token")
	}

	if time.Now().After(expiry) {
		panic("token is expired")
	}

	// make sure there are no other active sessions for the token
	if token.SessionActive {
		panic("Cannot have multiple sessions")
	}

	// check that the token is not revoked
	if token.Revoked {
		panic("Your access token was revoked")
	}

	// check the token is not archived
	if token.Archived {
		panic("token can no longer be used")
	}

	// start session
	session.UpdateSession(*tok, true)

	fmt.Println("Hosting a TLS API on port " + *port)
	fmt.Println("If you like this API, all donations are appreciated! https://paypal.me/carcraftz")

	mux := http.NewServeMux()
	mux.Handle("/", new(TLSHandler))

	srv := &http.Server{
		Addr: fmt.Sprintf("0.0.0.0:%s", *port),
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      mux,
	}

	go func() {
		if err := http.ListenAndServe(":"+string(*port), nil); err != nil {
			log.Fatalln("Error starting the HTTP server:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.

	c := make(chan os.Signal, 1)

	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	fmt.Println("shutdown server")

	session.UpdateSession(*tok, false)

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)

	os.Exit(0)
}

type TLSHandler struct {
}

func (h TLSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	// Ensure page URL header is provided
	pageURL := r.Header.Get("Poptls-Url")
	if pageURL == "" {
		http.Error(w, "ERROR: No Page URL Provided", http.StatusBadRequest)
		return
	}
	// Remove header to ignore later
	r.Header.Del("Poptls-Url")

	// Ensure user agent header is provided
	userAgent := r.Header.Get("User-Agent")
	if userAgent == "" {
		http.Error(w, "ERROR: No User Agent Provided", http.StatusBadRequest)
		return
	}

	//Handle Proxy (http://host:port or http://user:pass@host:port)
	proxy := r.Header.Get("Poptls-Proxy")
	if proxy != "" {
		r.Header.Del("Poptls-Proxy")
	}
	//handle redirects and timeouts
	redirectVal := r.Header.Get("Poptls-Allowredirect")
	allowRedirect := true
	if redirectVal != "" {
		if redirectVal == "false" {
			allowRedirect = false
		}
	}
	if redirectVal != "" {
		r.Header.Del("Poptls-Allowredirect")
	}
	timeoutraw := r.Header.Get("Poptls-Timeout")
	timeout, err := strconv.Atoi(timeoutraw)
	if err != nil {
		//default timeout of 6
		timeout = 6
	}
	if timeout > 60 {
		http.Error(w, "ERROR: Timeout cannot be longer than 60 seconds", http.StatusBadRequest)
		return
	}
	// Change JA3
	var tlsClient tls.ClientHelloID
	if strings.Contains(strings.ToLower(userAgent), "chrome") {
		tlsClient = tls.HelloChrome_Auto
	} else if strings.Contains(strings.ToLower(userAgent), "firefox") {
		tlsClient = tls.HelloFirefox_Auto
	} else {
		tlsClient = tls.HelloIOS_Auto
	}
	client, err := cclient.NewClient(tlsClient, proxy, allowRedirect, time.Duration(timeout))
	if err != nil {
		log.Fatal(err)
	}

	// Forward query params
	var addedQuery string
	for k, v := range r.URL.Query() {
		addedQuery += "&" + k + "=" + v[0]
	}

	endpoint := pageURL
	if len(addedQuery) != 0 {
		endpoint = pageURL + "?" + addedQuery
		if strings.Contains(pageURL, "?") {
			endpoint = pageURL + addedQuery
		} else if addedQuery != "" {
			endpoint = pageURL + "?" + addedQuery[1:]
		}
	}
	req, err := http.NewRequest(r.Method, ""+endpoint, r.Body)
	if err != nil {
		panic(err)
	}
	//master header order, all your headers will be ordered based on this list and anything extra will be appended to the end
	//if your site has any custom headers, see the header order chrome uses and then add those headers to this list
	masterheaderorder := []string{
		"host",
		"connection",
		"cache-control",
		"device-memory",
		"viewport-width",
		"rtt",
		"downlink",
		"ect",
		"sec-ch-ua",
		"sec-ch-ua-mobile",
		"sec-ch-ua-full-version",
		"sec-ch-ua-arch",
		"sec-ch-ua-platform",
		"sec-ch-ua-platform-version",
		"sec-ch-ua-model",
		"upgrade-insecure-requests",
		"user-agent",
		"accept",
		"sec-fetch-site",
		"sec-fetch-mode",
		"sec-fetch-user",
		"sec-fetch-dest",
		"referer",
		"accept-encoding",
		"accept-language",
		"cookie",
	}
	headermap := make(map[string]string)
	//TODO: REDUCE TIME COMPLEXITY (This code is very bad)
	headerorderkey := []string{}
	for _, key := range masterheaderorder {
		for k, v := range r.Header {
			lowercasekey := strings.ToLower(k)
			if key == lowercasekey {
				headermap[k] = v[0]
				headerorderkey = append(headerorderkey, lowercasekey)
			}
		}

	}
	for k, v := range req.Header {
		if _, ok := headermap[k]; !ok {
			headermap[k] = v[0]
			headerorderkey = append(headerorderkey, strings.ToLower(k))
		}
	}

	//ordering the pseudo headers and our normal headers
	req.Header = http.Header{
		http.HeaderOrderKey:  headerorderkey,
		http.PHeaderOrderKey: {":method", ":authority", ":scheme", ":path"},
	}
	//set our Host header
	u, err := url.Parse(endpoint)
	if err != nil {
		panic(err)
	}
	//append our normal headers
	for k := range r.Header {
		if k != "Content-Length" && !strings.Contains(k, "Poptls") {
			v := r.Header.Get(k)
			req.Header.Set(k, v)
		}
	}
	req.Header.Set("Host", u.Host)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[%s][%s][%s]\r\n", color.YellowString("%s", time.Now().Format("2012-11-01T22:08:41+00:00")), color.BlueString("%s", pageURL), color.RedString("Connection Failed"))
		hj, ok := w.(http.Hijacker)
		if !ok {
			panic(err)
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			panic(err)
		}
		if err := conn.Close(); err != nil {
			panic(err)
		}
		return
	}
	defer resp.Body.Close()

	//req.Close = true

	//forward response headers
	for k, v := range resp.Header {
		if k != "Content-Length" && k != "Content-Encoding" {
			for _, kv := range v {
				w.Header().Add(k, kv)
			}
		}
	}
	w.WriteHeader(resp.StatusCode)
	var status string
	if resp.StatusCode > 302 {
		status = color.RedString("%s", resp.Status)
	} else {
		status = color.GreenString("%s", resp.Status)
	}
	fmt.Printf("[%s][%s][%s]\r\n", color.YellowString("%s", time.Now().Format("2012-11-01T22:08:41+00:00")), color.BlueString("%s", pageURL), status)

	//forward decoded response body
	encoding := resp.Header["Content-Encoding"]
	body, err := ioutil.ReadAll(resp.Body)
	finalres := ""
	if err != nil {
		panic(err)
	}
	finalres = string(body)
	if len(encoding) > 0 {
		if encoding[0] == "gzip" {
			unz, err := gUnzipData(body)
			if err != nil {
				panic(err)
			}
			finalres = string(unz)
		} else if encoding[0] == "deflate" {
			unz, err := enflateData(body)
			if err != nil {
				panic(err)
			}
			finalres = string(unz)
		} else if encoding[0] == "br" {
			unz, err := unBrotliData(body)
			if err != nil {
				panic(err)
			}
			finalres = string(unz)
		} else {
			fmt.Println("UNKNOWN ENCODING: " + encoding[0])
			finalres = string(body)
		}
	} else {
		finalres = string(body)
	}
	if _, err := fmt.Fprint(w, finalres); err != nil {
		log.Println("Error writing body:", err)
	}
}

func gUnzipData(data []byte) (resData []byte, err error) {
	gz, _ := gzip.NewReader(bytes.NewReader(data))
	defer gz.Close()
	respBody, err := ioutil.ReadAll(gz)
	return respBody, err
}
func enflateData(data []byte) (resData []byte, err error) {
	zr, _ := zlib.NewReader(bytes.NewReader(data))
	defer zr.Close()
	enflated, err := ioutil.ReadAll(zr)
	return enflated, err
}
func unBrotliData(data []byte) (resData []byte, err error) {
	br := brotli.NewReader(bytes.NewReader(data))
	respBody, err := ioutil.ReadAll(br)
	return respBody, err
}
