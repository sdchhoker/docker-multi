package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"net/http"
	"os"
)
var db *sql.DB
var pool *redis.Pool

type Index struct {
	Index int `json:"index"`
}

func main() {
	connectToPostgres()
	createPool()
	ping()
	defer db.Close()
	setupServer()
}

func setupServer() {
	router := mux.NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusFound)
		w.Write([]byte("hello from server micro"))
	}).Methods("GET")

	router.HandleFunc("/values/all", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`SELECT number from values`)
		if err != nil {
			fmt.Println("err in getting values from db ", err)
		}
		defer rows.Close()
		results := make([] int, 0, 10)
		fmt.Println("in all")

		for rows.Next() {
			var number int
			rows.Scan(&number)
			results = append(results, number)
		}
		fmt.Println(results)
		json.NewEncoder(w).Encode(results)
	}).Methods("GET")

	router.HandleFunc("/values/current", func(w http.ResponseWriter, r *http.Request) {
		conn := pool.Get()
		reply, err := conn.Do("HGETALL", "values")
		if err != nil {
			fmt.Println("error in fetching ", err)
		}
		fmt.Println()
		val, err := redis.StringMap(reply, err)
		if err != nil {
			fmt.Println(err)
		}
		json.NewEncoder(w).Encode(val)
	}).Methods("GET")

	router.HandleFunc("/values", func(w http.ResponseWriter, r *http.Request) {
		var itx Index

		err := json.NewDecoder(r.Body).Decode(&itx)

		if err != nil {
			fmt.Println("error in reading body", err)
		}
		if itx.Index > 40 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("unable to process such large number"))
			return
		}
		conn := pool.Get()
		conn.Do("HSET", "values", itx.Index, "")
		conn.Do("PUBLISH", "message", itx.Index)
		fmt.Println(itx)
		_, err = db.Exec(`
			INSERT INTO values(number)
			VALUES ($1)
		`, itx.Index)
		if err != nil {
			fmt.Println("error in inserting values", err)
		}
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("calculating"))
	}).Methods("POST")
	http.ListenAndServe(":5000", handlers.CORS()(router))
}

func connectToPostgres () {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", os.Getenv("PGHOST"), os.Getenv("PGPORT"), os.Getenv("PGUSER"), os.Getenv("PGPASSWORD"), os.Getenv("PGDATABASE"))
	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		fmt.Println("not able to open connection", err)
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		fmt.Println("not able to ping db")
		panic(err)
	}
	fmt.Println("connected successfully")
	createTable()
}

func createTable() {
	result, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS values (
    	number integer
	)`)
	if err != nil {
		fmt.Println("error in creating table", err)
		panic(err)
	}
	fmt.Println("result ", result)
}

func createPool(){
	pool = &redis.Pool{
		MaxIdle: 10,
		MaxActive: 10,
		Dial: func() (redis.Conn, error) {

			conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")))
			if err != nil {
				fmt.Println("not able to connect to redis", err)
				os.Exit(1)
			}
			return conn, err
		},
	}
}

func ping() {
	conn := pool.Get()
	pong, err := redis.String(conn.Do("PING"))
	if err != nil {
		fmt.Println("not able to ping", err)
		os.Exit(1)
	}
	fmt.Println("pinged redis", pong)
}
