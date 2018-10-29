package events

import (
	"testing"

	"github.com/containerd/typeurl"
	eventsapi "github.com/ehazlett/stellar/api/services/events/v1"
)

const (
	testURI = "stellar.test/Test"
)

func init() {
	typeurl.Register(&Test{}, testURI)
}

type Test struct {
	Data string
}

func TestMarshalEvent(t *testing.T) {
	tt := &Test{
		Data: "123",
	}
	a, err := MarshalEvent(tt)
	if err != nil {
		t.Fatal(err)
	}
	if a.TypeUrl != testURI {
		t.Fatalf("expected uri %s; received %s", testURI, a.TypeUrl)
	}
}

func TestUnmarshalEvent(t *testing.T) {
	tt := &Test{
		Data: "456",
	}
	data, err := MarshalEvent(tt)
	if err != nil {
		t.Fatal(err)
	}

	v, err := UnmarshalEvent(&eventsapi.Message{Subject: "test", Data: data})
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := v.(*Test); !ok {
		t.Fatalf("expected type Test; received %+v", v)
	}
}
