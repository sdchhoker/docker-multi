package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"os"
	"strconv"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	pool := createPool()
	ping(pool)
	subscribe(pool)
	wg.Wait()
}

func fib(num int64) int {
	if num <= 2 {
		return 1
	}
	return fib(num-1) + fib(num-2)
}

func subscribe(pool *redis.Pool) {
	conn := pool.Get()
	psc := redis.PubSubConn{Conn: conn}
	psc.PSubscribe("message")
	LOOP: for {
		switch msg := psc.Receive().(type) {
		case redis.Message:
			key := string(msg.Data)
			fmt.Printf("Message: %s %s\n", msg.Channel, msg.Data)
			value, err := strconv.ParseInt(key, 10, 64)
			if err != nil {
				fmt.Println("not able to parse value", err)
				goto LOOP
			}
			conn := pool.Get()
			conn.Do("HSET", "values", msg.Data, fib(value))

		case redis.Subscription:
			fmt.Printf("Subscription: %s %s %d\n", msg.Kind, msg.Channel, msg.Count)
			if msg.Count == 0 {
				return
			}
		case error:
			fmt.Printf("error: %v\n", msg)
			return
		}
	}
}

func createPool() *redis.Pool {
	pool := &redis.Pool{
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
	return pool
}

func ping(pool *redis.Pool) {
	conn := pool.Get()
	pong, err := redis.String(conn.Do("PING"))
	if err != nil {
		fmt.Println("not able to ping", err)
		os.Exit(1)
	}
	fmt.Println(pong)
}
