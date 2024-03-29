package utils

import (
	"fmt"
	"testing"
)

func TestNewUUID(t *testing.T) {
	uid, err := NewUUID()
	if err != nil {
		t.Errorf("Error of generate new id:%v",err)
	}
	fmt.Printf("New id is: %s\n",uid)
}

func TestNewID(t *testing.T) {
	id := NewID()
	t.Log("Xid:" + id)
}

func TestNewStreamID(t *testing.T) {
	for i:= 1; i < 50; i++ {
		id, _:= NewStreamID()
		t.Log("Stream_id:" + id)
	}
}
