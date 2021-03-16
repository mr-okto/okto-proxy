package repeater

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"proxy-server/internal/database"
	"strconv"
	"strings"
)

var db *sql.DB
var httpClient *http.Client

func GetPrevRequests(writer http.ResponseWriter, req *http.Request) {
	err := database.WriteRequests(db, writer)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
	}
}

func GetRequest(writer http.ResponseWriter, req *http.Request) {
	reqIDStr, ok := mux.Vars(req)["id"]
	reqId, err := strconv.ParseInt(reqIDStr, 10, 32)
	if !ok || err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	req, err = database.GetRequestDetails(db, int(reqId))
	if errors.Is(err, database.ErrReqNotFound) {
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	host := strings.Split(req.URL.Host, ":")[0]
	req.RequestURI = fmt.Sprintf("%s://%s%s", req.URL.Scheme, host, req.URL.Path)
	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.Write(requestDump)
	writer.Write([]byte("\r\n"))
}

func RepeatRequest(writer http.ResponseWriter, req *http.Request) {
	reqIDStr, ok := mux.Vars(req)["id"]
	reqId, err := strconv.ParseInt(reqIDStr, 10, 32)
	if !ok || err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	req, err = database.GetRequestDetails(db, int(reqId))
	if errors.Is(err, database.ErrReqNotFound) {
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	res, err := httpClient.Do(req)
	if err != nil && !errors.Is(err, http.ErrUseLastResponse) {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	for key, values := range res.Header {
		for _, v := range values {
			writer.Header().Add(key, v)
		}
	}
	writer.WriteHeader(res.StatusCode)
	io.Copy(writer, res.Body)
}

func Start() {
	var err error
	db, err = database.OpenDB()
	if err != nil {
		log.Fatalln("Unable to open database: ", err)
	}
	httpClient = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	r := mux.NewRouter()
	r.HandleFunc("/requests", GetPrevRequests).Methods("GET")
	r.HandleFunc("/requests/{id}", GetRequest).Methods("GET")
	r.HandleFunc("/repeat/{id}", RepeatRequest).Methods("GET")
	err = http.ListenAndServe(":8000", r)
	if err != nil {
		log.Fatal(err)
		return
	}
}
