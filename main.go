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

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type response struct {
	Frequency int `json:"frequency"`
}

var addr string
var result response
var frequencyGauge = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "accelerator_signal_frequency_hertz",
		Help: "main frequency of linac",
	}, []string{"tip"},
)

func api(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(result)
	if err != nil {
		log.Println("Cannot marshall")
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
	if err != nil {
		log.Println("Cannot write to api")
	}
}

func main() {
	reg := prometheus.NewRegistry()
	reg.MustRegister(frequencyGauge)
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
	log.Println("Server is on ", addr)

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
				frequencyGauge.WithLabelValues("main frequency of linac").Set(float64(result.Frequency))
				//log.Println(result.Frequency)
				time.Sleep(1 * time.Second)
			}
			time.Sleep(5 * time.Second)
		}
	}()
	http.HandleFunc("/api", api)
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	log.Println("api running on :8080/api")
	log.Println("metrics running on :8080/metrics")
	http.ListenAndServe(":8080", nil)
	wg.Wait()
	log.Println("ended")
}
