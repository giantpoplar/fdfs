package cluster

import (
	"testing"
	//  "time"
)

func TestFixString(t *testing.T) {
	s := "helloworld"
	r := fixString(s, 5)
	if r != "hello" {
		t.Error("test fix string fail")
	}
	r = fixString(s, 11)
	if r != "helloworld"+string(0) {
		t.Error("test fix string fail")
	}
}

func Test_stripString(t *testing.T) {
	b := []byte("hello")
	b = append(b, byte(0))
	s := stripString(string(b))
	println(s, len(s))

}
