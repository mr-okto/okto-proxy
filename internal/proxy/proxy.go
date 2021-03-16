package proxy

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"proxy-server/internal/certificates"
	"proxy-server/internal/database"
)

var db *sql.DB

func handleHTTP(w http.ResponseWriter, req *http.Request) {
	log.Println(req.Method, req.RequestURI)
	req.Header.Del("Proxy-Connection")
	req.RequestURI = req.URL.Path
	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	go database.SaveRequest(db, req, "http", reqBody)
	response, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer response.Body.Close()
	for key, values := range w.Header() {
		for _, v := range values {
			response.Header.Add(key, v)
		}
	}
	w.WriteHeader(response.StatusCode)
	io.Copy(w, response.Body)
}

func tunnel(dest io.WriteCloser, src io.ReadCloser) {
	defer dest.Close()
	defer src.Close()
	io.Copy(dest, src)
}

func copyTunnel(dest io.WriteCloser, src io.ReadCloser) {
	defer dest.Close()
	defer src.Close()
	buf := new(bytes.Buffer)
	writer := io.MultiWriter(dest, buf)
	io.Copy(writer, src)
	req, err := http.ReadRequest(bufio.NewReader(buf))
	if err == nil && req != nil {
		log.Println(fmt.Sprintf("%s https://%s%s", req.Method, req.Host, req.URL.Path))
		req.URL.Host = req.Host + ":443"
		go database.SaveRequest(db, req, "https", nil)
	}
}

func handleHTTPS(w http.ResponseWriter, req *http.Request) {
	cert, err := certificates.GetCert(req.Host)
	if err != nil {
		log.Println(err)
	}
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacker error", http.StatusInternalServerError)
		return
	}
	answerConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = answerConn.Write([]byte("HTTP/1.0 200 Connection established\r\n\r\n"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*cert},
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return certificates.GetCert(info.ServerName)
		},
	}
	destConn, err := tls.Dial("tcp", req.Host, tlsConfig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	servConn := tls.Server(answerConn, tlsConfig)
	err = servConn.Handshake()
	go tunnel(servConn, destConn)
	go copyTunnel(destConn, servConn)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		handleHTTPS(w, r)
	} else {
		handleHTTP(w, r)
	}
}

func Start() {
	err := certificates.LoadRootCert()
	if err != nil {
		log.Fatalln("Unable to load certificate: ", err)
	}
	db, err = database.InitDB()
	if err != nil {
		log.Fatalln("Unable to init database: ", err)
	}
	server := &http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(Handler),
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	log.Printf("Launching Proxy at port %s", server.Addr)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
		return
	}
}
