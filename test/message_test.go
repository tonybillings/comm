package test

import (
	"testing"
	"tonysoft.com/comm/internal/node"
)

type SomeStruct struct {
	SomeTextField    string
	SomeNumericField float64
	SomeMapField     map[string]interface{}
}

func TestMessageSerializationNoPayload(t *testing.T) {
	msg := node.NewMessage[any](8001, ":8002", nil)

	msgBytes, err := msg.ToBytes()
	if err != nil {
		t.Error(err)
		return
	}

	msgCopy, err := node.MessageFromBytes[any](msgBytes)
	if err != nil {
		t.Error(err)
		return
	}

	if msgCopy.FromNode() != msg.FromNode() {
		t.Errorf("msgCopy.FromNode (%s) != msg.FromNode (%s)", msgCopy.FromNode(), msg.FromNode())
		return
	}

	if msgCopy.ID() != msg.ID() {
		t.Errorf("msgCopy.ID (%d) != msg.ID (%d)", msgCopy.ID(), msg.ID())
		return
	}

	if msgCopy.Status() != msg.Status() {
		t.Errorf("msgCopy.Status (%d) != msg.Status (%d)", msgCopy.Status(), msg.Status())
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
}

func TestMessageSerializationByteArrayPayload(t *testing.T) {
	data := []byte{1, 2, 3, 4}
	msg := node.NewMessage[[]byte](8001, ":8002", &data)

	msgBytes, err := msg.ToBytes()
	if err != nil {
		t.Error(err)
		return
	}

	msgCopy, err := node.MessageFromBytes[[]byte](msgBytes)
	if err != nil {
		t.Error(err)
		return
	}

	if msgCopy.FromNode() != msg.FromNode() {
		t.Errorf("msgCopy.FromNode (%s) != msg.FromNode (%s)", msgCopy.FromNode(), msg.FromNode())
		return
	}

	if msgCopy.ID() != msg.ID() {
		t.Errorf("msgCopy.ID (%d) != msg.ID (%d)", msgCopy.ID(), msg.ID())
		return
	}

	if msgCopy.Status() != msg.Status() {
		t.Errorf("msgCopy.Status (%d) != msg.Status (%d)", msgCopy.Status(), msg.Status())
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
}

func TestMessageSerializationStructPayload(t *testing.T) {
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

	msgCopy, err := node.MessageFromBytes[SomeStruct](msgBytes)
	if err != nil {
		t.Error(err)
		return
	}

	if msgCopy.FromNode() != msg.FromNode() {
		t.Errorf("msgCopy.FromNode (%s) != msg.FromNode (%s)", msgCopy.FromNode(), msg.FromNode())
		return
	}

	if msgCopy.ID() != msg.ID() {
		t.Errorf("msgCopy.ID (%d) != msg.ID (%d)", msgCopy.ID(), msg.ID())
		return
	}

	if msgCopy.Status() != msg.Status() {
		t.Errorf("msgCopy.Status (%d) != msg.Status (%d)", msgCopy.Status(), msg.Status())
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
}
