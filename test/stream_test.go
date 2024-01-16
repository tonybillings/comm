package test

import (
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"tonysoft.com/comm/internal/node"
	"tonysoft.com/comm/pkg/stream"
)

func TestStringStream(t *testing.T) {
	const threadCount = 5
	const wordCount = 500
	const maxWordLength = 5000

	startingRoutineCount := runtime.NumGoroutine()

	var wg sync.WaitGroup
	wg.Add(threadCount)
	for i := 0; i < threadCount; i++ {
		go func() {
			defer wg.Done()

			var wordsStream []byte
			words := make([]string, wordCount)
			for j := 0; j < wordCount; j++ {
				word := GetRandomCString(10 + int(rand.Uint32()%(maxWordLength-10)))
				words[j] = string(word[:len(word)-1])
				wordsStream = append(wordsStream, word...)
			}

			r := stream.NewDataReader(wordsStream)

			var wordsCopy []string
			for word := range stream.String(r) {
				wordsCopy = append(wordsCopy, word)
			}

			for j := 0; j < wordCount; j++ {
				if words[j] != wordsCopy[j] {
					t.Errorf("unexpected word found (expected %s, have %s)",
						words[j][:5]+"..."+words[j][len(words[j])-5:], wordsCopy[j][:5]+"..."+wordsCopy[j][len(wordsCopy[j])-5:])
					return
				}
			}
		}()
	}

	wg.Wait()

	finishingRoutineCount := runtime.NumGoroutine()
	if finishingRoutineCount > startingRoutineCount {
		t.Errorf("unexpected thread count (expected <=%d, have %d)", startingRoutineCount, finishingRoutineCount)
	}
}

func TestMessageStreamNoPayload(t *testing.T) {
	msg := node.NewMessage[any](8001, ":8002", nil)

	msgBytes, err := msg.ToBytes()
	if err != nil {
		t.Error(err)
		return
	}

	r := stream.NewDataReader(msgBytes)

	testPassed := false
	for msgCopy := range node.NewMessageStream[any](r) {
		if msgCopy.FromNode() != msg.FromNode() {
			t.Errorf("msgCopy.FromNode (%s) != msg.FromNode (%s)", msgCopy.FromNode(), msg.FromNode())
			return
		}

		if msgCopy.ID() != msg.ID() {
			t.Errorf("msgCopy.ID (%d) != msg.ID (%d)", msgCopy.ID(), msg.ID())
			return
		}

		if msgCopy.Status() != node.MessageReceived {
			t.Errorf("msgCopy.Status (%d) != node.MessageReceived", msgCopy.Status())
			return
		}

		if msgCopy.SentOn() != msg.SentOn() {
			t.Errorf("msgCopy.SentOn (%v) != msg.SentOn (%v)", msgCopy.SentOn(), msg.SentOn())
			return
		}

		if msgCopy.ReceivedOn() != msg.ReceivedOn() {
			t.Errorf("msgCopy.ReceivedOn (%v) != msg.ReceivedOn (%v)", msgCopy.ReceivedOn(), msg.ReceivedOn())
			return
		}

		testPassed = true
	}

	if !testPassed {
		t.Error("test failed")
		return
	}
}

func TestMessageStreamByteArrayPayload(t *testing.T) {
	data := []byte{1, 2, 3, 4}
	msg := node.NewMessage[[]byte](8001, ":8002", &data)

	msgBytes, err := msg.ToBytes()
	if err != nil {
		t.Error(err)
		return
	}

	r := stream.NewDataReader(msgBytes)

	testPassed := false
	for msgCopy := range node.NewMessageStream[[]byte](r) {
		if msgCopy.FromNode() != msg.FromNode() {
			t.Errorf("msgCopy.FromNode (%s) != msg.FromNode (%s)", msgCopy.FromNode(), msg.FromNode())
			return
		}

		if msgCopy.ID() != msg.ID() {
			t.Errorf("msgCopy.ID (%d) != msg.ID (%d)", msgCopy.ID(), msg.ID())
			return
		}

		if msgCopy.Status() != node.MessageReceived {
			t.Errorf("msgCopy.Status (%d) != node.MessageReceived", msgCopy.Status())
			return
		}

		if msgCopy.SentOn() != msg.SentOn() {
			t.Errorf("msgCopy.SentOn (%v) != msg.SentOn (%v)", msgCopy.SentOn(), msg.SentOn())
			return
		}

		if msgCopy.ReceivedOn() != msg.ReceivedOn() {
			t.Errorf("msgCopy.ReceivedOn (%v) != msg.ReceivedOn (%v)", msgCopy.ReceivedOn(), msg.ReceivedOn())
			return
		}

		for i := 0; i < len(data); i++ {
			if (*msgCopy.Data)[i] != (*msg.Data)[i] {
				t.Errorf("(*msgCopy.Data)[i] (%d) != (*msg.Data)[i] (%d)", (*msgCopy.Data)[i], (*msg.Data)[i])
				return
			}
		}

		testPassed = true
	}

	if !testPassed {
		t.Error("test failed")
		return
	}
}

func TestMessageStreamStructPayload(t *testing.T) {
	data := &SomeStruct{
		SomeTextField:    "abc123",
		SomeNumericField: 1.234,
		SomeMapField: map[string]any{
			"key1": "xyz",
			"key2": 555.555,
		},
	}

	msg := node.NewMessage[SomeStruct](8001, ":8002", data)

	msgBytes, err := msg.ToBytes()
	if err != nil {
		t.Error(err)
		return
	}

	r := stream.NewDataReader(msgBytes)

	testPassed := false
	for msgCopy := range node.NewMessageStream[SomeStruct](r) {
		if msgCopy.FromNode() != msg.FromNode() {
			t.Errorf("msgCopy.FromNode (%s) != msg.FromNode (%s)", msgCopy.FromNode(), msg.FromNode())
			return
		}

		if msgCopy.ID() != msg.ID() {
			t.Errorf("msgCopy.ID (%d) != msg.ID (%d)", msgCopy.ID(), msg.ID())
			return
		}

		if msgCopy.Status() != node.MessageReceived {
			t.Errorf("msgCopy.Status (%d) != node.MessageReceived", msgCopy.Status())
			return
		}

		if msgCopy.SentOn() != msg.SentOn() {
			t.Errorf("msgCopy.SentOn (%v) != msg.SentOn (%v)", msgCopy.SentOn(), msg.SentOn())
			return
		}

		if msgCopy.ReceivedOn() != msg.ReceivedOn() {
			t.Errorf("msgCopy.ReceivedOn (%v) != msg.ReceivedOn (%v)", msgCopy.ReceivedOn(), msg.ReceivedOn())
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

		testPassed = true
	}

	if !testPassed {
		t.Error("test failed")
		return
	}
}
