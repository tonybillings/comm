package test

import (
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"tonysoft.com/comm/pkg/node"
)

func TestNodeCommNoPayload(t *testing.T) {
	startingRoutineCount := runtime.NumGoroutine()

	func() {
		cfg1 := node.NewConfig(":9001")
		cfg2 := node.NewConfig(":9002")

		n1, err := node.New[any](cfg1)
		if err != nil {
			t.Error(err)
			return
		}

		n2, err := node.New[any](cfg2)
		if err != nil {
			t.Error(err)
			return
		}

		err = n1.Start()
		if err != nil {
			t.Error(err)
			return
		}
		defer n1.Stop()

		err = n2.Start()
		if err != nil {
			t.Error(err)
			return
		}
		defer n2.Stop()

		msg, err := n1.Send(n2.Config().Address, nil)
		if err != nil {
			t.Error(err)
			return
		}

		time.Sleep(100 * time.Millisecond)

		msgCopy := <-n2.Recv()

		if msgCopy.ID() != msg.ID() {
			t.Errorf("msgCopy.ID (%d) != msg.ID (%d)", msgCopy.ID(), msg.ID())
			return
		}

		if msgCopy.Status() != node.MessageReceived {
			t.Errorf("msgCopy.Status (%d) != node.MessageReceived", msgCopy.Status())
			return
		}
	}()

	time.Sleep(time.Second)

	finishingRoutineCount := runtime.NumGoroutine()
	if finishingRoutineCount > startingRoutineCount {
		t.Errorf("unexpected thread count (expected %d, have %d)", startingRoutineCount, finishingRoutineCount)
	}
}

func TestNodeCommByteArrayPayload(t *testing.T) {
	startingRoutineCount := runtime.NumGoroutine()

	func() {
		cfg1 := node.NewConfig(":9001")
		cfg2 := node.NewConfig(":9002")

		n1, err := node.New[[]byte](cfg1)
		if err != nil {
			t.Error(err)
			return
		}

		n2, err := node.New[[]byte](cfg2)
		if err != nil {
			t.Error(err)
			return
		}

		err = n1.Start()
		if err != nil {
			t.Error(err)
			return
		}
		defer n1.Stop()

		err = n2.Start()
		if err != nil {
			t.Error(err)
			return
		}
		defer n2.Stop()

		data := []byte{1, 2, 3, 4}

		msg, err := n1.Send(n2.Config().Address, &data)
		if err != nil {
			t.Error(err)
			return
		}

		time.Sleep(100 * time.Millisecond)

		msgCopy := <-n2.Recv()

		if msgCopy.ID() != msg.ID() {
			t.Errorf("msgCopy.ID (%d) != msg.ID (%d)", msgCopy.ID(), msg.ID())
			return
		}

		if msgCopy.Status() != node.MessageReceived {
			t.Errorf("msgCopy.Status (%d) != node.MessageReceived", msgCopy.Status())
			return
		}

		dataCopy := *msgCopy.Data
		for i := 0; i < len(data); i++ {
			if dataCopy[i] != data[i] {
				t.Errorf("unexpected byte received in message data (expected %d, have %d)", data[i], dataCopy[i])
				return
			}
		}
	}()

	time.Sleep(time.Second)

	finishingRoutineCount := runtime.NumGoroutine()
	if finishingRoutineCount > startingRoutineCount {
		t.Errorf("unexpected thread count (expected %d, have %d)", startingRoutineCount, finishingRoutineCount)
	}
}

func TestNodeCommStructPayload(t *testing.T) {
	startingRoutineCount := runtime.NumGoroutine()

	func() {
		cfg1 := node.NewConfig(":9001")
		cfg2 := node.NewConfig(":9002")

		n1, err := node.New[SomeStruct](cfg1)
		if err != nil {
			t.Error(err)
			return
		}

		n2, err := node.New[SomeStruct](cfg2)
		if err != nil {
			t.Error(err)
			return
		}

		err = n1.Start()
		if err != nil {
			t.Error(err)
			return
		}
		defer n1.Stop()

		err = n2.Start()
		if err != nil {
			t.Error(err)
			return
		}
		defer n2.Stop()

		data := &SomeStruct{
			SomeTextField:    "abc123",
			SomeNumericField: 1.234,
			SomeMapField: map[string]any{
				"key1": "xyz",
				"key2": 555.555,
			},
		}

		msg, err := n1.Send(n2.Config().Address, data)
		if err != nil {
			t.Error(err)
			return
		}

		time.Sleep(100 * time.Millisecond)

		msgCopy := <-n2.Recv()

		if msgCopy.ID() != msg.ID() {
			t.Errorf("msgCopy.ID (%d) != msg.ID (%d)", msgCopy.ID(), msg.ID())
			return
		}

		if msgCopy.Status() != node.MessageReceived {
			t.Errorf("msgCopy.Status (%d) != node.MessageReceived", msgCopy.Status())
			return
		}

		if msgCopy.Data.SomeTextField != msg.Data.SomeTextField {
			t.Errorf("msgCopy.Data.SomeTextField (%s) != msg.Data.SomeTextField (%s)", msgCopy.Data.SomeTextField, msg.Data.SomeTextField)
			return
		}

		if msgCopy.Data.SomeNumericField != msg.Data.SomeNumericField {
			t.Errorf("msgCopy.Data.SomeNumericField (%f) != msg.Data.SomeNumericField (%f)", msgCopy.Data.SomeNumericField, msg.Data.SomeNumericField)
			return
		}

		if msgCopy.Data.SomeMapField["key1"].(string) != msg.Data.SomeMapField["key1"].(string) {
			t.Errorf("msgCopy.Data.SomeMapField['key1'] (%s) != msg.Data.SomeMapField['key1'] (%s)", msgCopy.Data.SomeMapField["key1"].(string), msg.Data.SomeMapField["key1"].(string))
			return
		}

		if msgCopy.Data.SomeMapField["key2"].(float64) != msg.Data.SomeMapField["key2"].(float64) {
			t.Errorf("msgCopy.Data.SomeMapField['key2'] (%f) != msg.Data.SomeMapField['key2'] (%f)", msgCopy.Data.SomeMapField["key2"].(float64), msg.Data.SomeMapField["key2"].(float64))
			return
		}
	}()

	time.Sleep(time.Second)

	finishingRoutineCount := runtime.NumGoroutine()
	if finishingRoutineCount > startingRoutineCount {
		t.Errorf("unexpected thread count (expected %d, have %d)", startingRoutineCount, finishingRoutineCount)
	}
}

func TestNodeCommLargePayload(t *testing.T) {
	startingRoutineCount := runtime.NumGoroutine()
	data := GetRandomCString(500000000) // 500 MB

	func() {
		cfg1 := node.NewConfig(":9001")
		cfg2 := node.NewConfig(":9002")

		n1, err := node.New[[]byte](cfg1)
		if err != nil {
			t.Error(err)
			return
		}

		n2, err := node.New[[]byte](cfg2)
		if err != nil {
			t.Error(err)
			return
		}

		err = n1.Start()
		if err != nil {
			t.Error(err)
			return
		}
		defer n1.Stop()

		err = n2.Start()
		if err != nil {
			t.Error(err)
			return
		}
		defer n2.Stop()

		msg, err := n1.Send(n2.Config().Address, &data)
		if err != nil {
			t.Error(err)
			return
		}

		time.Sleep(time.Second)

		msgCopy := <-n2.Recv()

		if msgCopy.ID() != msg.ID() {
			t.Errorf("msgCopy.ID (%d) != msg.ID (%d)", msgCopy.ID(), msg.ID())
			return
		}

		if msgCopy.Status() != node.MessageReceived {
			t.Errorf("msgCopy.Status (%d) != node.MessageReceived", msgCopy.Status())
			return
		}

		dataCopy := *msgCopy.Data
		for i := 0; i < len(data); i++ {
			if dataCopy[i] != data[i] {
				t.Errorf("unexpected byte received in message data (expected %d, have %d)", data[i], dataCopy[i])
				return
			}
		}
	}()

	time.Sleep(time.Second)

	finishingRoutineCount := runtime.NumGoroutine()
	if finishingRoutineCount > startingRoutineCount {
		t.Errorf("unexpected thread count (expected %d, have %d)", startingRoutineCount, finishingRoutineCount)
	}
}

func TestMultiNodeComm(t *testing.T) {
	startingRoutineCount := runtime.NumGoroutine()

	const nodeCount = 50
	const startingPort = 9630
	const messageCount = 500

	var skippedCount atomic.Uint32
	var passCount atomic.Uint32
	var msgSentCount atomic.Uint32
	var rcptRecvCount atomic.Uint32

	nodes := make([]node.Node[string], nodeCount)
	for i := 0; i < nodeCount; i++ {
		replyPort := startingPort + i

		cfg := node.NewConfig(fmt.Sprintf(":%d", replyPort))
		n, err := node.New[string](cfg)
		if err != nil {
			t.Error(err)
			return
		}

		err = n.Start()
		if err != nil {
			t.Error(err)
			return
		}

		nodes[i] = n
	}

	var wg sync.WaitGroup
	wg.Add(nodeCount)
	run := func(wg *sync.WaitGroup, n node.Node[string]) {
		defer wg.Done()

		replyPort, _ := strconv.Atoi(strings.Split(n.Config().Address, ":")[1])

		recvChan := make(chan bool)

		// Send messages
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < messageCount; j++ {
				destPort := (int(rand.Uint32()) % nodeCount) + startingPort
				if destPort == replyPort {
					skippedCount.Add(1)
					continue
				}

				destAddr := fmt.Sprintf(":%d", destPort)
				data := fmt.Sprintf("reply:%d dest:%d", replyPort, destPort)

				_, err := n.Send(destAddr, &data)
				if err != nil {
					t.Error(err)
					return
				}
				msgSentCount.Add(1)
			}
		}()

		// Verify message receipts
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-recvChan:
					return
				case rcpt, ok := <-n.Status():
					if !ok {
						return
					}

					if rcpt.Status() == node.MessageReceived {
						rcptRecvCount.Add(1)
					}
				}
			}
		}()

		// Receive messages
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-recvChan:
					return
				case msg, ok := <-n.Recv():
					if !ok {
						return
					}

					expectedData := fmt.Sprintf("reply:%s dest:%d", strings.Split(msg.FromNode(), ":")[1], replyPort)
					data := *msg.Data
					if data != expectedData {
						t.Errorf("unexpected message received (expected '%s', have '%s')", expectedData, data)
						break
					} else {
						passCount.Add(1)
					}
				}
			}
		}()

		time.Sleep(5 * time.Second)

		close(recvChan)
		n.Stop()
	}

	for _, n := range nodes {
		go run(&wg, n)
	}

	wg.Wait()
	time.Sleep(time.Second)

	expectedPassCount := uint32(nodeCount*messageCount - int(skippedCount.Load()))
	if passCount.Load() != expectedPassCount {
		t.Errorf("unexpected pass count (expected %d, have %d)", expectedPassCount, passCount.Load())
		return
	}

	if msgSentCount.Load() != rcptRecvCount.Load() {
		t.Errorf("unexpected receipt count (expected %d, have %d)", msgSentCount.Load(), rcptRecvCount.Load())
		return
	}

	finishingRoutineCount := runtime.NumGoroutine()
	if finishingRoutineCount > startingRoutineCount {
		t.Errorf("unexpected thread count (expected %d, have %d)", startingRoutineCount, finishingRoutineCount)
		return
	}
}
