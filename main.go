package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type response struct {
	Frequency int  `json:"frequency"`
	Status    bool `json:"status,omitempty"`
}

var (
	addr   string
	result response
	mu     sync.RWMutex

	frequencyGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "accelerator_signal_frequency_hertz",
			Help: "main frequency of linac",
		},
		[]string{"tip"},
	)
)

func api(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	b, err := json.Marshal(result)
	mu.RUnlock()

	if err != nil {
		http.Error(w, "Cannot marshal", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func main() {
	reg := prometheus.NewRegistry()
	reg.MustRegister(frequencyGauge)

	flag.StringVar(&addr, "addr", "127.0.0.1:52550", "Address of tcp server")
	flag.Parse()

	log.Println("Server is on", addr)

	// Клиент для чтения данных с сервера
	go func() {
		for {
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				log.Printf("cannot dial tcp %s: %v", addr, err)
				time.Sleep(5 * time.Second)
				continue
			}

			decoder := json.NewDecoder(conn)

			conn.SetReadDeadline(time.Now().Add(10 * time.Second))

			if err := decoder.Decode(&result); err != nil {
				if errors.Is(err, io.EOF) {
					log.Println("server closed connection after sending data (normal)")
				} else if errors.Is(err, io.ErrUnexpectedEOF) {
					log.Println("partial data received — server closed connection prematurely")
				} else {
					log.Printf("decode error: %v", err)
				}
				mu.Lock()
				frequencyGauge.WithLabelValues("main").Set(float64(math.NaN()))
				mu.Unlock()

			} else {
				mu.Lock()
				frequencyGauge.WithLabelValues("main").Set(float64(result.Frequency))
				mu.Unlock()
				//log.Printf("parsed: %+v", result)
			}

			conn.Close()
			time.Sleep(1 * time.Second)
		}
	}()

	http.HandleFunc("/api", api)
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	log.Println("api running on :8084/api")
	log.Println("metrics running on :8084/metrics")

	if err := http.ListenAndServe(":8084", nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
