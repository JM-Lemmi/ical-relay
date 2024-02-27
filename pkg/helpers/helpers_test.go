package helpers

import (
	"net/http"
	"os"
	"testing"
)

var w http.ResponseWriter

func OkOrHttpErrorResponseBad(t *testing.T) {
	handle := OkOrHttpError[[]byte](w, "test.bad", "read", "not there")
	a := handle(os.ReadFile("test.bad"))
	// a should be nil
	if a != nil {
		t.Error("Expected nil, got", a)
	}
}

func OkOrHttpErrorResponseGood(t *testing.T) {
	handle := OkOrHttpError[[]byte](w, "test.txt", "read", "not there")
	a := handle(os.ReadFile("test.txt"))
	if a == nil {
		t.Error("Expected nil, got", a)
	}
}

func TestOkOrPanicPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	handle := OkOrPanic[[]byte]("test.bad", "read", "not there")
	a := handle(os.ReadFile("test.bad"))
	if a != nil {
		t.Error("Expected nil, got", a)
	}
}

func TestOkOrPanicOk(t *testing.T) {
	handle := OkOrPanic[[]byte]("test.txt", "read", "not there")
	a := handle(os.ReadFile("test.txt"))
	if a == nil {
		t.Error("Expected nil, got", a)
	}
}
