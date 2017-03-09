package json

import (
	"bytes"
	"testing"
)

type encodeTest struct {
	in  interface{}
	out string
}

var encodeTests = []encodeTest{
	{in: true, out: "true"},
	{in: 3, out: "3"},
	{in: "Hello", out: `"Hello"`},
	{in: S1{X: 10, Y: 20}, out: `{"X":10,"Y":20}`},
	{in: map[string]int{"X": 10, "Y": 20}, out: `{"X":10,"Y":20}`},
	{in: []string{"X", "Y", "Z"}, out: `["X","Y","Z"]`},
}

func TestEncode(t *testing.T) {
	for i, tst := range encodeTests {
		b := new(bytes.Buffer)
		e := NewEncoder(b)
		if err := e.Encode(tst.in); err != nil {
			t.Errorf("#%d: %v", i, err)
		}
		if b.String() != tst.out {
			t.Errorf("#%d: mismatch %v want %v", i, b.String(), tst.out)
		}
	}
}
