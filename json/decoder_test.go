package json

import (
	"reflect"
	"strings"
	"testing"
)

type S1 struct {
	X int
	Y int
}

type S2 struct {
	ID      string
	Content string
}

type decodeTest struct {
	in  string
	out interface{}
	ptr interface{}
}

var decodeTests = []decodeTest{
	{in: `true`, ptr: new(bool), out: true},
	{in: `false`, ptr: new(bool), out: false},
	{in: `1`, ptr: new(int), out: 1},
	{in: `3.14`, ptr: new(float64), out: 3.14},
	{in: `"x"`, ptr: new(string), out: "x"},
	{in: `null`, ptr: new(*int), out: (*int)(nil)},
	{in: `{"X": 4, "Y": 7}`, ptr: new(S1), out: S1{X: 4, Y: 7}},
	{
		in:  `{"ID": "1192", "Content": "Hello, world"}`,
		ptr: new(S2),
		out: S2{ID: "1192", Content: "Hello, world"},
	},
	{
		in:  `{"X": 4, "Y": 7}`,
		ptr: new(map[string]int),
		out: map[string]int{"X": 4, "Y": 7},
	},
	{
		in:  `{"ID": "1192", "Content": "Hello, world"}`,
		ptr: new(map[string]string),
		out: map[string]string{"ID": "1192", "Content": "Hello, world"},
	},
}

func TestDecode(t *testing.T) {
	for i, tst := range decodeTests {
		d := NewDecoder(strings.NewReader(tst.in))
		v := reflect.New(reflect.TypeOf(tst.ptr).Elem())
		if err := d.Decode(v.Interface()); err != nil {
			t.Errorf("#%d: %v", i, err)
		}
		if !reflect.DeepEqual(v.Elem().Interface(), tst.out) {
			t.Errorf("#%d: mismatch %v want %v", i, v.Elem().Interface(), tst.out)
		}
	}
}
