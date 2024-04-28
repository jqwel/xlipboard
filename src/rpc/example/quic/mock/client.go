package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	quic "github.com/quic-go/quic-go"
	quichttp "github.com/quic-go/quic-go/http3"
)

func main() {
	roundTripper := &quichttp.RoundTripper{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		QuicConfig: &quic.Config{},
	}

	client := &http.Client{
		Transport: roundTripper,
	}

	var count int64
	for range 1000000 {
		go func() {
			data := []byte("your binary data here")
			req, err := http.NewRequestWithContext(context.Background(), "POST", "https://localhost:3216/demo/echo", bytes.NewBuffer(data))
			if err != nil {
				return
			}
			req.Header.Set("Content-Type", "application/octet-stream")
			req.Header.Set("Authorization", "Bearer YOUR_API_TOKEN")
			resp, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Server's reply %d: %s\n", atomic.AddInt64(&count, 1), body)
		}()
	}

	//resp, err := client.Get("https://localhost:3216/demo/tiles")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer resp.Body.Close()
	//
	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	log.Fatal(err)
	//}

	fmt.Printf("Server's reply: %s\n", "body")
	time.Sleep(time.Second * 20)
}
