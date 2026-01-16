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
						conn.Write([]byte("hello\n"))
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
				log.Print(string(b[:n]))
				time.Sleep(1 * time.Second)
			}
			time.Sleep(5 * time.Second)
		}
	}()
	wg.Wait()
	log.Println("ended")
}
