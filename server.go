package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"reflect"
	"strconv"
)

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer next.ServeHTTP(w, r)
		log.Println(r.RequestURI, r.Header)
	})
}

func errorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %+v", err)
				w.WriteHeader(http.StatusInternalServerError)
				if w.Header().Get("Content-Type") == "application/json" {
					w.Write([]byte(`{"error": "500 Internal Server Error"}`))
				} else {
					f, _ := os.Open(path.Join("static", "error.html"))
					io.Copy(w, f)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tem, err := template.ParseFiles(path.Join(homeDir, staticDir, "index.html"))
	handleError(err)
	err = tem.Execute(w, nil)
	handleError(err)
}

func notFoundHanlder(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{message: "Not found."}`))
}

type ArticleInfo struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

func getResult(rows *sql.Rows, returnType interface{}, args ...string) interface{} {
	t := reflect.TypeOf(returnType)
	arr := reflect.MakeSlice(reflect.SliceOf(t), 0, 10)
	for rows.Next() {
		fields := make([]reflect.Value, len(args))
		for i := range args {
			fields[i] = reflect.ValueOf(returnType).Elem().FieldByName(args[i]).Addr()
		}
		err := rows.Scan(fields)
		handleError(err)

		arr = reflect.Append(arr, reflect.ValueOf(returnType))
	}
	return arr.Interface()
}

func autoCompeleteApiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	k := r.URL.Query().Get("k")
	if len(k) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{error: "Invalid parameter."}`))
		return
	}
	db, err := sql.Open("mysql", dataSource)
	handleError(err)
	defer db.Close()

	arr := make([]interface{}, 0, 5)
	rows, err := db.Query("SELECT `id`, `title` FROM `test` WHERE MATCH (`title`)"+
		"AGAINST(? IN NATURAL LANGUAGE MODE) LIMIT 5;", k)
	for rows.Next() {
		info := struct {
			Id    int    `json:"id"`
			Title string `json:"title"`
		}{}
		err = rows.Scan(&info.Id, &info.Title)
		handleError(err)

		arr = append(arr, info)
	}

	data, err := json.Marshal(arr)
	handleError(err)
	w.Write(data)
}
func queryApiHandler(w http.ResponseWriter, r *http.Request) {
	const (
		pageList  = 10
		threshold = 5
	)
	w.Header().Add("Content-Type", "application/json")

	q := r.URL.Query().Get("q")
	p := r.URL.Query().Get("p")

	if p == "" {
		p = "0"
	}
	page, err := strconv.Atoi(p)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{error: "Invalid parameter."}`))
		return
	}

	offset := page * pageList

	db, err := sql.Open("mysql", dataSource)
	handleError(err)
	defer db.Close()

	count := 0
	rows, err := db.Query("SELECT COUNT(`id`) FROM `test` WHERE "+
		"MATCH(`title`, `text`) AGAINST (? IN NATURAL LANGUAGE MODE) > ?;", q, threshold)
	handleError(err)

	rows.Next()
	err = rows.Scan(&count)
	handleError(err)

	rows, err = db.Query("SELECT `id`, `title`, SUBSTRING(`text`, 100, 200) FROM `test` "+
		"WHERE MATCH(`title`, `text`) AGAINST (? IN NATURAL LANGUAGE MODE) > ? "+
		"LIMIT 10 OFFSET ?;", q, threshold, offset)
	handleError(err)
	defer rows.Close()

	arr := make([]ArticleInfo, 0, 10)

	for rows.Next() {
		info := ArticleInfo{}
		err = rows.Scan(&info.Id, &info.Title, &info.Text)
		handleError(err)
		arr = append(arr, info)
	}
	err = rows.Err()
	handleError(err)

	returnVal, err := json.Marshal(struct {
		Data  []ArticleInfo `json:"data"`
		Count int           `json:"count"`
	}{arr, count})
	handleError(err)

	w.Write(returnVal)
}

const (
	host         = "35.236.173.141"
	port         = "3306"
	defaultDb    = "demo"
	password     = "michael@tw.ibm.com"
	homeDir      = "/Users/changheng/Desktop/Go/src/github.com/qq52184962/sqlgen/"
	staticPrefix = "/static/"
	staticDir    = "static"
	apiPrefix    = "/api/v1/"
)

var dataSource = fmt.Sprintf(
	"root:%s@tcp(%s:%s)/%s?timeout=90s&collation=utf8mb4_unicode_ci",
	password,
	host,
	port,
	defaultDb,
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	router := mux.NewRouter()
	router.StrictSlash(true)
	router.Use(logMiddleware, errorMiddleware)

	router.PathPrefix(staticPrefix).Handler(
		http.StripPrefix(staticPrefix,
			http.FileServer(http.Dir(path.Join(homeDir, staticDir))),
		))

	apiRputer := router.PathPrefix(apiPrefix).Subrouter()
	apiRputer.HandleFunc("/search", queryApiHandler)
	apiRputer.HandleFunc("/suggest", autoCompeleteApiHandler)

	router.HandleFunc("/", indexHandler)
	router.HandleFunc("/index", indexHandler)
	router.NotFoundHandler = http.HandlerFunc(notFoundHanlder)

	go func() {
		log.Fatal(http.ListenAndServe("localhost:8080", router))
	}()
	<-c
}
