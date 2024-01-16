package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"tonysoft.com/comm/pkg/server"
	"tonysoft.com/comm/pkg/stream"
)

var (
	localAddress        = "00:00:00:00:00:00"
	localPort    uint16 = 5
)

// Unless you have multiple Bluetooth adapters installed, after starting the
// echo server you will need to run the client defined in ../client_test on
// another computer.  If you do have multiple adapters and would like to run
// both the client and server on the same computer, then you can bind the
// server to a specific one by setting the localAddress variable the address
// of the adapter you want the server to use, etc.
func echoServer() {
	cfg := server.NewConfig(localAddress, localPort)
	s, err := server.New(cfg)
	if err != nil {
		panic(err)
	}

	go func() {
		for conn := range s.Accept() {
			go func(c server.Connection) {
				fmt.Printf("client connected: %v\n", c.RemoteAddress())
				for str := range stream.String(c) {
					fmt.Printf("[%s] received: %s\n", c.RemoteAddress(), str)
					_, e := c.Write([]byte(str))
					if e != nil {
						panic(e)
					}
				}
			}(conn)
		}
	}()

	go func() {
		for e := range s.Errors() {
			panic(e)
		}
	}()

	err = s.Start()
	if err != nil {
		panic(err)
	}
	defer s.Stop()

	fmt.Println("server started...")

	var wg sync.WaitGroup
	wg.Add(1)
	intChan := make(chan os.Signal, 1)
	signal.Notify(intChan, os.Interrupt)
	go func() {
		<-intChan
		log.Println("interrupt received, stopping server...")
		signal.Stop(intChan)
		wg.Done()
	}()

	wg.Wait()
}

func main() {
	echoServer()
}
