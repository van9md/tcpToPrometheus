package main

import (
	"bufio"
	"flag"
	"log"
	"net"
	"sync"
	"time"
)

func connect(addr string) (*net.Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	log.Println("Connected to ", conn)
	return &conn, nil
}

func main() {
	var wg sync.WaitGroup
	var addr string
	flag.StringVar(&addr, "addr", "127.0.0.1:52550", "Address of tcp server, leave empty to simulate server")

	ln, err := net.Listen("tcp", ":52550")
	if err != nil {
		log.Println("Cannot up test server")
	}
	//test server
	go func(ln net.Listener) {
		defer wg.Done()
		wg.Add(1)
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Println("Cannot accept on test tcp")
			}
			go func(conn net.Conn) {
				for {
					conn.Write([]byte("hello\n"))
					time.Sleep(1 * time.Second)
					if err != nil {
						log.Println("breaking")
						break
					}
				}
			}(conn)
		}
	}(ln)

	time.Sleep(2*time.Second)
	conn, err := net.Dial("tcp", "127.0.0.1:52550")
	if err != nil {
		log.Println("cannot dial tcp")
	}
	go func(conn net.Conn) {
		defer wg.Done()
		wg.Add(1)
		buf := bufio.NewReader(conn)
		for {
			data, err := buf.ReadString('\n')
			if err != nil {
				break
			}
			log.Print(data)
		}
	}(conn)
	wg.Wait()
	log.Println("ended")
}
