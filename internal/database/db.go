package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

const dbname = "proxy-server.db"

var ErrReqNotFound = fmt.Errorf("request not found")

func createRequestTable(db *sql.DB) error {
	createTableSQL := `create table if not exists request(
    id integer primary key autoincrement,
    method text,
	scheme text,
    url_host text,
    url_path text,
    headers text,
    host text,
    uri text,
    body text
	);`
	statement, err := db.Prepare(createTableSQL)
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	if err != nil {
		return err
	}
	return nil
}

func OpenDB() (*sql.DB, error) {
	return sql.Open("sqlite3", dbname)
}

func InitDB() (*sql.DB, error) {
	file, err := os.Create(dbname)
	if err != nil {
		return nil, fmt.Errorf("unable to create sqlite3 database: %s", err.Error())
	}
	err = file.Close()
	if err != nil {
		return nil, fmt.Errorf("file operation error: %s", err.Error())
	}
	db, err := OpenDB()
	if err != nil {
		return nil, fmt.Errorf("unable to open sqlite3 database: %s", err.Error())
	}
	err = createRequestTable(db)
	if err != nil {
		return nil, fmt.Errorf("unable to create request table: %s", err.Error())
	}
	return db, nil
}

func SaveRequest(db *sql.DB, r* http.Request, scheme string, rBody []byte) error {
	query := `
	INSERT INTO Request(method, scheme, url_host, url_path, headers, body, host, uri) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8);`
	headers, err := json.Marshal(r.Header)
	if err == nil && rBody == nil {
		rBody, err = ioutil.ReadAll(r.Body)
	}
	if err != nil {
		return err
	}
	_, err = db.Exec(query, r.Method, scheme, r.URL.Host, r.URL.Path, headers, string(rBody),
		r.Host, r.RequestURI)
	return err
}

func WriteRequests(db *sql.DB, writer http.ResponseWriter) error {
	rows, err := db.Query(`
	SELECT id, method, scheme, host, uri
	FROM Request ORDER BY id DESC`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var method string
		var scheme string
		var host string
		var uri string
		rows.Scan(&id, &method, &scheme, &host, &uri)
		reqInfo := fmt.Sprintf("<p><a href=\"http://localhost:8000/requests/%d\">#%d:</a> ",
			id, id)
		reqInfo += fmt.Sprintf("%s %s://%s%s</p>\r\n", method, scheme, host, uri)
		writer.Write([]byte(reqInfo))
	}
	return err
}

func GetRequestDetails(db *sql.DB, reqId int) (* http.Request, error) {
	row, err := db.Query(`
	SELECT method, scheme, url_host, url_path, headers, body
	FROM Request WHERE id = $1`, reqId)
	if err != nil {
		return nil, err
	}
	defer row.Close()
	var method string
	var scheme string
	var urlHost string
	var urlPath string
	var headers string
	var body string
	if row.Next() == false {
		return nil, ErrReqNotFound
	}
	err = row.Scan(&method, &scheme, &urlHost, &urlPath, &headers, &body)
	if err == sql.ErrNoRows {
		return nil, ErrReqNotFound
	}
	if err != nil {
		return nil, err
	}
	address := fmt.Sprintf("%s://%s%s", scheme, urlHost, urlPath)
	request, err := http.NewRequest(method, address, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(headers), &request.Header)
	return request, nil
}
