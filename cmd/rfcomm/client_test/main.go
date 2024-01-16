package main

import (
	"fmt"
	"math/rand"
	"time"
	"tonysoft.com/comm/pkg/client"
)

var (
	remoteAddress        = "94:B8:6D:91:06:D5"
	remotePort    uint16 = 5
)

// This assumes you have the echo server defined in ../server_test running
func echoClient() {
	cfg := client.NewConfig(remoteAddress, remotePort)
	c, err := client.New(cfg)
	if err != nil {
		panic(err)
	}

	err = c.Start()
	if err != nil {
		panic(err)
	}
	defer func() {
		e := c.Stop()
		if e != nil {
			panic(e)
		}
	}()

	data := getRandomCString(50)

	count, err := c.Write(data)
	if err != nil {
		panic(err)
	}
	if count != len(data) {
		panic(fmt.Errorf("expected to write %d bytes, wrote %d instead", len(data), count))
	}
	fmt.Printf("sent: %s\n", data)

	time.Sleep(time.Second)

	response := make([]byte, len(data))
	count, err = c.Read(response)
	if err != nil {
		panic(err)
	}
	if count != len(data)-1 {
		panic(fmt.Errorf("expected to read %d bytes, read %d instead", len(response), count))
	}

	if string(response) != string(data) {
		panic(fmt.Errorf("expected to read '%s', read '%s' instead", string(data), string(response)))
	}

	fmt.Printf("received: %s\n", string(response))
}

func getRandomCString(length int) []byte {
	data := make([]byte, length+1)
	for i := 0; i < length; i++ {
		data[i] = byte(65 + (rand.Uint32() % 26))
	}
	return data
}

func main() {
	fmt.Println("running echo test...")
	echoClient()
	fmt.Println("test succeeded")
}
