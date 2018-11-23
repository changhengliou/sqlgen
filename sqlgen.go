package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

type Doc struct {
	Id    string `json:"id"`
	Url   string `json:"url"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

func getFilePathFromDir(dir string) []string {
	arr := make([]string, 0)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if file.IsDir() {
			for _, f := range getFilePathFromDir(path.Join(dir, file.Name())) {
				arr = append(arr, f)
			}
		} else {
			arr = append(arr, path.Join(dir, file.Name()))
		}
	}
	return arr
}

func escape(sql string) string {
	dest := make([]byte, 0, 2*len(sql))
	var escape byte
	for i := 0; i < len(sql); i++ {
		c := sql[i]
		escape = 0
		switch c {
		case 0: /* Must be escaped for 'mysql' */
			escape = '0'
			break
		case '\n': /* Must be escaped for logs */
			escape = 'n'
			break
		case '\r':
			escape = 'r'
			break
		case '\\':
			escape = '\\'
			break
		case '\'':
			escape = '\''
			break
		case '"':
			escape = '"'
			break
		case '\032':
			escape = 'Z'
		}

		if escape != 0 {
			dest = append(dest, '\\', escape)
		} else {
			dest = append(dest, c)
		}
	}

	return string(dest)
}

func sqlBuilder(docs []Doc, sql *[]string, mutex *sync.Mutex) {
	for _, doc := range docs {
		id, err := strconv.Atoi(doc.Id)
		if err != nil {
			panic(err)
		}
		mutex.Lock()
		*sql = append(*sql, fmt.Sprintf(`(%d, "%s", "%s", "%s")`,
			id, doc.Url, doc.Title, escape(doc.Text)))
		mutex.Unlock()
	}
}

func saveToDb(sqlString []string, db *sql.DB) {
	sqlStr := "INSERT INTO test (`id`, `url`, `title`, `text`) VALUES "

	sqlStr += strings.Join(sqlString, ",") + ";"
	_, err := db.Exec(sqlStr)
	if err != nil {
		panic(err)
	}
	fmt.Println("commit...")
}

func main() {
	var (
		DB_HOST    = "localhost"
		DB_PORT    = "3306"
		DEFAULT_DB = "test"
		DB_PASSWORD,
		ENTRY_DIR  string
	)
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		currArg := args[i]

		if currArg == "-h" || currArg == "-P" || currArg == "-p" || currArg == "-t" {
			i++
			if i >= len(args) {
				panic("insufficient argument")
			}
		}
		if currArg == "-h" {
			DB_HOST = args[i]
		} else if currArg == "-P" {
			DB_PORT = args[i]
		} else if currArg == "-p" {
			DB_PASSWORD = args[i]
		} else if currArg == "-t" {
			DEFAULT_DB = args[i]
		} else {
			ENTRY_DIR = currArg
		}
	}
	dbSource := fmt.Sprintf("root:%s@tcp(%s:%s)"+
		"/%s?timeout=90s&collation=utf8mb4_unicode_ci",
		DB_PASSWORD, DB_HOST, DB_PORT, DEFAULT_DB)
	fmt.Println(dbSource)

	files := getFilePathFromDir(ENTRY_DIR)
	sqlString := make([]string, 0)
	mutex := &sync.Mutex{}
	readDone := false
	wg := &sync.WaitGroup{}
	wg.Add(3)

	go func() {
		for _, filePath := range files {
			arr := make([]Doc, 0)

			data, err := ioutil.ReadFile(filePath)
			if err != nil {
				panic(err)
			}
			decoder := json.NewDecoder(bytes.NewBuffer(data))
			result := Doc{}
			fmt.Println(filePath)
			for {
				err := decoder.Decode(&result)
				if err == io.EOF {
					break
				}
				if err != nil {
					panic(err)
				}
				arr = append(arr, result)
			}
			sqlBuilder(arr, &sqlString, mutex)
		}
		readDone = true
	}()

	save := func() {
		db, err := sql.Open("mysql", dbSource)
		if err != nil {
			panic(err)
		}
		defer db.Close()

		tmp := make([]string, 0)
		for !(readDone && len(sqlString) == 0) {
			mutex.Lock()
			if len(sqlString) > 100 {
				tmp = sqlString[:100]
				sqlString = sqlString[100:]
				mutex.Unlock()
				saveToDb(tmp, db)
			} else if readDone {
				tmp = sqlString[:]
				sqlString = sqlString[:0]
				mutex.Unlock()
				saveToDb(tmp, db)
			} else {
				mutex.Unlock()
				runtime.Gosched()
			}
		}
		wg.Done()
	}
	for i := 0; i < 3; i++ {
		go save()
	}
	wg.Wait()
}
