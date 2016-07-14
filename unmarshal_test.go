package jsonapi

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func assertEqual(t *testing.T, expect interface{}, v interface{}, descr ...string) {
	if !reflect.DeepEqual(v, expect) {
		_, fname, lineno, ok := runtime.Caller(1)
		if !ok {
			fname, lineno = "<UNKNOWN>", -1
		}
		var des string
		if len(descr) == 0 {
			des = fmt.Sprintf("Expected: %#v\nReceived: %#v", expect, v)
		} else {
			des = strings.Join(descr, ", ")
		}
		t.Errorf("FAIL: %s:%d\n%s", fname, lineno, des)
	}
}

func assertNil(t *testing.T, v interface{}, descr ...string) {
	if !reflect.ValueOf(v).IsValid() {
		return
	}
	if !reflect.ValueOf(v).IsNil() {
		_, fname, lineno, ok := runtime.Caller(1)
		if !ok {
			fname, lineno = "<UNKNOWN>", -1
		}
		var des string
		if len(descr) == 0 {
			des = fmt.Sprintf("Expected to be nil but received: %#v\n", v)
		} else {
			des = strings.Join(descr, ", ")
		}
		t.Errorf("FAIL: %s:%d\n%s", fname, lineno, des)
	}
}

type testStruct1 struct {
	ID         uint64                 `jsonapi:"id,test-structs1"`
	String     string                 `jsonapi:"attr,string"`
	Bool       bool                   `jsonapi:"attr,bool"`
	Map        map[string]interface{} `jsonapi:"attr,map"`
	Slice      []int                  `jsonapi:"attr,slice"`
	Sub        testSub                `jsonapi:"attr,sub"`
	WontUpdate string                 `jsonapi:"attr,wont-update,readonly"`
	Excluded   string
}

func TestUnmarshal(t *testing.T) {
	s := testStruct1{}

	req := `{"data":{"id":"100","type":"test-structs1","attributes":{"String":"str","bool":true,"map":{"a":"1","b":"2","c":{"a1":"11"}},"slice":[1,2,3],"sub":{"country":"CTR","city":"DT"},"wont-update":"readonly string","Excluded":"no"}}}`

	err := Unmarshal([]byte(req), &s)
	assertNil(t, err)
	assertEqual(t, "str", s.String)
	assertEqual(t, true, s.Bool)
	wantMap := map[string]interface{}{"a": "1", "b": "2", "c": map[string]interface{}{"a1": "11"}}
	assertEqual(t, wantMap, s.Map)
	wantSlice := []int{1, 2, 3}
	assertEqual(t, wantSlice, s.Slice)
	wantSub := testSub{Country: "CTR", City: "DT"}
	assertEqual(t, wantSub, s.Sub)
	assertEqual(t, "", s.WontUpdate)
	assertEqual(t, "", s.Excluded)
}

func TestUnmarshalWithChanges(t *testing.T) {
	s := testStruct1{
		ID:         100,
		String:     "str",
		Bool:       false,
		Map:        map[string]interface{}{"a": "1", "b": "2", "c": map[string]interface{}{"a1": "11"}},
		Slice:      []int{1, 2, 3},
		Sub:        testSub{Country: "CTR", City: "DT"},
		WontUpdate: "never changed",
		Excluded:   "dont include this",
	}

	req := `{"data":{"id":"100","type":"test-structs1","attributes":{"String":"str1","bool":true,"map":{"a":"1","b":"22","d":"4","c":{"a1":"111"}},"slice":[1,2,3,4],"sub":{"country":"CTR","city":"DT1"},"wont-update":"change me","Excluded":"change me"}}}`

	changes, err := UnmarshalWithChanges([]byte(req), &s)
	// fmt.Println(changes)
	assertNil(t, err)

	assertEqual(t, Change{Field: "string", Cur: "str", New: "str1"}, changes.Find("string"))
	assertEqual(t, Change{Field: "bool", Cur: "false", New: "true"}, changes.Find("bool"))
	assertEqual(t, Change{Field: "map.b", Cur: "2", New: "22"}, changes.Find("map.b"))
	assertEqual(t, Change{Field: "map.c.a1", Cur: "11", New: "111"}, changes.Find("map.c.a1"))
	assertEqual(t, Change{Field: "map.d", Cur: "", New: "4"}, changes.Find("map.d"))
	assertEqual(t, Change{Field: "slice", Cur: "[1 2 3]", New: "[1 2 3 4]"}, changes.Find("slice"))
	assertEqual(t, Change{Field: "sub.City", Cur: "DT", New: "DT1"}, changes.Find("sub.City"))
	assertEqual(t, 7, len(changes))
}

func BenchmarkUnmarshalPlain(b *testing.B) {
	s := testStruct1{
		ID:         100,
		String:     "str",
		Bool:       false,
		Map:        map[string]interface{}{"a": "1", "b": "2", "c": map[string]interface{}{"a1": "11"}},
		Slice:      []int{1, 2, 3},
		Sub:        testSub{Country: "CTR", City: "DT"},
		WontUpdate: "never changed",
		Excluded:   "dont include this",
	}
	req := []byte(`{"data":{"id":"100","type":"test-structs1","attributes":{"String":"str1","bool":true,"map":{"a":"1","b":"22","c":{"a1":"111"},"d":"4"},"slice":[1,2,3,4],"sub":{"country":"CTR","city":"DT1"},"wont-update":"change me","Excluded":"change me"}}}`)
	for i := 0; i < b.N; i++ {
		Unmarshal(req, &s)
	}
}

func BenchmarkUnmarshalChanges(b *testing.B) {
	s := testStruct1{
		ID:         100,
		String:     "str",
		Bool:       false,
		Map:        map[string]interface{}{"a": "1", "b": "2", "c": map[string]interface{}{"a1": "11"}},
		Slice:      []int{1, 2, 3},
		Sub:        testSub{Country: "CTR", City: "DT"},
		WontUpdate: "never changed",
		Excluded:   "dont include this",
	}
	req := []byte(`{"data":{"id":"100","type":"test-structs1","attributes":{"String":"str1","bool":true,"map":{"a":"1","b":"22","c":{"a1":"111"},"d":"4"},"slice":[1,2,3,4],"sub":{"country":"CTR","city":"DT1"},"wont-update":"change me","Excluded":"change me"}}}`)
	for i := 0; i < b.N; i++ {
		UnmarshalWithChanges(req, &s)
	}
}

func BenchmarkUnmarshalChanges1(b *testing.B) {
	s := testStruct1{
		ID:         100,
		String:     "str",
		Bool:       false,
		Sub:        testSub{Country: "CTR", City: "DT"},
		WontUpdate: "never changed",
		Excluded:   "dont include this",
	}
	req := []byte(`{"data":{"id":"100","type":"test-structs1","attributes":{"String":"str1","bool":true,"sub":{"country":"CTR","city":"DT1"},"wont-update":"change me","Excluded":"change me"}}}`)
	for i := 0; i < b.N; i++ {
		UnmarshalWithChanges(req, &s)
	}
}
