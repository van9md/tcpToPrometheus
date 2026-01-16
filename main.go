package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type response struct {
	Frequency int `json:"frequency"`
}

var addr string
var result response

func api(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(result)
	if err != nil {
		log.Println("Cannot marshall")
	}
	w.Write(b)
	if err != nil {
		log.Println("Cannot write to api")
	}
}
func main() {
	var wg sync.WaitGroup
	flag.StringVar(&addr, "addr", "127.0.0.1:52550", "Address of tcp server, leave empty to simulate server")

	//test server
	if addr == "127.0.0.1:52550" {
		ln, err := net.Listen("tcp", ":52550")
		if err != nil {
			log.Println("Cannot up test server")
		}
		go func(ln net.Listener) {
			defer wg.Done()
			wg.Add(1)
			for {
				conn, err := ln.Accept()
				if err != nil {
					log.Println("Cannot accept on test tcp")
				}
				for {
					conn.Write([]byte(`{"frequency":10}`))
					time.Sleep(1 * time.Second)
					if err != nil {
						log.Println("breaking")
						break
					}
				}
			}
		}(ln)
	}

	time.Sleep(1 * time.Second)
	go func() {
		defer wg.Done()
		b := make([]byte, 4096)
		for {
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				log.Println("cannot dial tcp")
				time.Sleep(5 * time.Second)
				continue
			}
			buf := bufio.NewReader(conn)

			for {
				n, err := buf.Read(b)
				if err != nil {
					log.Println("read error:", err)
					conn.Close()
					break
				}
				err = json.Unmarshal(b[:n], &result)
				if err != nil {
					log.Println("unmarshal error")
					break
				}
				log.Println(result.Frequency)
				time.Sleep(1 * time.Second)
			}
			time.Sleep(5 * time.Second)
		}
	}()
	http.HandleFunc("/api", api)
	http.ListenAndServe(":8080", nil)
	wg.Wait()
	log.Println("ended")
}
