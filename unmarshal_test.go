package jsonapi

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestUnquote(t *testing.T) {
	v := unquote([]byte(`""`))
	assertEqual(t, []byte(""), v)
	v = unquote([]byte(`"qwe"`))
	assertEqual(t, []byte("qwe"), v)
}

type testStructUnmarshaler struct {
	ID   uint64 `jsonapi:"id,test-structs1"`
	Name string `jsonapi:"attr,name"`
}

func (t *testStructUnmarshaler) UnmarshalJSONAPI(b []byte) error {
	t.Name = "custom name"
	return nil
}

func TestUnmarshaler(t *testing.T) {
	s := testStructUnmarshaler{}
	req := `{"data":{"id":"100","type":"test-structs1","attributes":{"Name":"John"}}}`

	err := Unmarshal([]byte(req), &s)
	assertNil(t, err)
	assertEqual(t, "custom name", s.Name)
}

type testStructAfterUnmarshaler struct {
	ID    uint64 `jsonapi:"id,test-structs1"`
	Name  string `jsonapi:"attr,name"`
	Email string `jsonapi:"attr,email"`
}

func (t *testStructAfterUnmarshaler) AfterUnmarshalJSONAPI() error {
	t.Email = "changed"
	return nil
}

func TestAfterUnmarshaler(t *testing.T) {
	s := testStructAfterUnmarshaler{}
	req := `{"data":{"id":"100","type":"test-structs1","attributes":{"name":"John"}}}`

	err := Unmarshal([]byte(req), &s)
	assertNil(t, err)
	assertEqual(t, "John", s.Name)
	assertEqual(t, "changed", s.Email)
}

type testStruct1 struct {
	ID         uint64                 `jsonapi:"id,test-structs1"`
	StringName string                 `jsonapi:"attr,string"`
	Bool       bool                   `jsonapi:"attr,bool"`
	Map        map[string]interface{} `jsonapi:"attr,map"`
	Slice      []int                  `jsonapi:"attr,slice"`
	Sub        testSub                `jsonapi:"attr,sub"`
	SubPointer *testSub               `jsonapi:"attr,subP"`
	Int        int                    `jsonapi:"attr,int"`
	IntStr     int                    `jsonapi:"attr,intstr,string"`
	WontUpdate string                 `jsonapi:"attr,wont-update,readonly"`
	Excluded   string
}

func TestUnmarshalStatic(t *testing.T) {
	s := testStruct1{}

	req := `{"data":{"id":"100","type":"test-structs1","attributes":{"string":"str","int":1,"intstr":"2","bool":true,"map":{"a":"1","b":"2","c":{"a1":"11"}},"slice":[1,2,3],"sub":{"country":"CTR","city":"DT"},"subP":{"country":"CTR1","city":"DT1"},"wont-update":"readonly string","Excluded":"no"}}}`

	err := Unmarshal([]byte(req), &s)
	assertNil(t, err)
	assertEqual(t, "str", s.StringName)
	assertEqual(t, true, s.Bool)
	wantMap := map[string]interface{}{"a": "1", "b": "2", "c": map[string]interface{}{"a1": "11"}}
	assertEqual(t, wantMap, s.Map)
	wantSlice := []int{1, 2, 3}
	assertEqual(t, wantSlice, s.Slice)
	wantSub := testSub{Country: "CTR", City: "DT"}
	assertEqual(t, wantSub, s.Sub)
	wantSubP := testSub{Country: "CTR1", City: "DT1"}
	assertEqual(t, &wantSubP, s.SubPointer)
	assertEqual(t, 1, s.Int)
	assertEqual(t, 2, s.IntStr)
	assertEqual(t, "", s.WontUpdate)
	assertEqual(t, "", s.Excluded)
}

func TestUnmarshalWithChanges(t *testing.T) {
	s := testStruct1{
		ID:         100,
		StringName: "str",
		Bool:       false,
		Map:        map[string]interface{}{"a": "1", "b": "2", "c": map[string]interface{}{"a1": "11"}},
		Slice:      []int{1, 2, 3},
		Sub:        testSub{Country: "CTR", City: "DT"},
		WontUpdate: "never changed",
		Excluded:   "dont include this",
	}

	req := `{"data":{"id":"100","type":"test-structs1","attributes":{"string":"str1","bool":true,"map":{"a":"1","b":"22","d":"4","c":{"a1":"111"}},"slice":[1,2,3,4],"sub":{"country":"CTR","city":"DT1"},"wont-update":"change me","Excluded":"change me"}}}`

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

type scopeTest struct {
	ID   uint64 `jsonapi:"id,test-structs"`
	S1   string `jsonapi:"attr,s1"`
	S2   string `jsonapi:"attr,s2" scope:"2"`
	S3   string `jsonapi:"attr,s3" scope:"3"`
	Both string `jsonapi:"attr,both" scope:"2,3"`
}

func TestUnmarshalWithScope(t *testing.T) {
	b := []byte(`{"data":{"id":"100","type":"test-structs","attributes":{"s1":"c1","s2":"c2","s3":"c3","both":"c4"}}}`)

	s := scopeTest{ID: 100, S1: "v1", S2: "v2", S3: "v3", Both: "v4"}
	assert.NoError(t, UnmarshalWithScope(b, &s, ""))
	assert.Equal(t, "c1", s.S1)
	assert.Equal(t, "c2", s.S2)
	assert.Equal(t, "c3", s.S3)
	assert.Equal(t, "c4", s.Both)

	s = scopeTest{ID: 100, S1: "v1", S2: "v2", S3: "v3", Both: "v4"}
	assert.NoError(t, UnmarshalWithScope(b, &s, "2"))
	assert.Equal(t, "c1", s.S1)
	assert.Equal(t, "c2", s.S2)
	assert.Equal(t, "v3", s.S3)
	assert.Equal(t, "c4", s.Both)

	s = scopeTest{ID: 100, S1: "v1", S2: "v2", S3: "v3", Both: "v4"}
	assert.NoError(t, UnmarshalWithScope(b, &s, "3"))
	assert.Equal(t, "c1", s.S1)
	assert.Equal(t, "v2", s.S2)
	assert.Equal(t, "c3", s.S3)
	assert.Equal(t, "c4", s.Both)

	s = scopeTest{ID: 100, S1: "v1", S2: "v2", S3: "v3", Both: "v4"}
	assert.NoError(t, UnmarshalWithScope(b, &s, "4"))
	assert.Equal(t, "c1", s.S1)
	assert.Equal(t, "v2", s.S2)
	assert.Equal(t, "v3", s.S3)
	assert.Equal(t, "v4", s.Both)
}

func BenchmarkUnmarshalPlain(b *testing.B) {
	s := testStruct1{
		ID:         100,
		StringName: "str",
		Bool:       false,
		Map:        map[string]interface{}{"a": "1", "b": "2", "c": map[string]interface{}{"a1": "11"}},
		Slice:      []int{1, 2, 3},
		Sub:        testSub{Country: "CTR", City: "DT"},
		WontUpdate: "never changed",
		Excluded:   "dont include this",
	}
	req := []byte(`{"data":{"id":"100","type":"test-structs1","attributes":{"string":"str1","bool":true,"map":{"a":"1","b":"22","c":{"a1":"111"},"d":"4"},"slice":[1,2,3,4],"sub":{"country":"CTR","city":"DT1"},"wont-update":"change me","Excluded":"change me"}}}`)
	for i := 0; i < b.N; i++ {
		Unmarshal(req, &s)
	}
}

func BenchmarkUnmarshalChanges(b *testing.B) {
	s := testStruct1{
		ID:         100,
		StringName: "str",
		Bool:       false,
		Map:        map[string]interface{}{"a": "1", "b": "2", "c": map[string]interface{}{"a1": "11"}},
		Slice:      []int{1, 2, 3},
		Sub:        testSub{Country: "CTR", City: "DT"},
		WontUpdate: "never changed",
		Excluded:   "dont include this",
	}
	req := []byte(`{"data":{"id":"100","type":"test-structs1","attributes":{"string":"str1","bool":true,"map":{"a":"1","b":"22","c":{"a1":"111"},"d":"4"},"slice":[1,2,3,4],"sub":{"country":"CTR","city":"DT1"},"wont-update":"change me","Excluded":"change me"}}}`)
	for i := 0; i < b.N; i++ {
		UnmarshalWithChanges(req, &s)
	}
}

func BenchmarkUnmarshalChanges1(b *testing.B) {
	s := testStruct1{
		ID:         100,
		StringName: "str",
		Bool:       false,
		Sub:        testSub{Country: "CTR", City: "DT"},
		WontUpdate: "never changed",
		Excluded:   "dont include this",
	}
	req := []byte(`{"data":{"id":"100","type":"test-structs1","attributes":{"string":"str1","bool":true,"sub":{"country":"CTR","city":"DT1"},"wont-update":"change me","Excluded":"change me"}}}`)
	for i := 0; i < b.N; i++ {
		UnmarshalWithChanges(req, &s)
	}
}
