package jsonapi

import "testing"

type testSub struct {
	Country string `json:"country"`
	City    string `json:"city"`
}

type testStruct struct {
	ID       uint64  `jsonapi:"id,test-structs"`
	Name     string  `jsonapi:"attr,name"`
	Address  testSub `jsonapi:"attr,address-at"`
	Excluded string
}

type testStructNonAPI struct {
	ID       uint64
	Name     string
	Address  testSub
	Excluded string
}

type testStructMarshaler struct {
	ID   uint64 `jsonapi:"id,test-structs"`
	Name string `jsonapi:"attr,name"`
}

func (t *testStructMarshaler) MarshalJSONAPI() error {
	t.Name = "changed"
	return nil
}

func TestMarshaler(t *testing.T) {
	s := testStructMarshaler{ID: 100}
	want := `{"id":"100","type":"test-structs","attributes":{"name":"changed"}}`

	res, err := Marshal(&s)
	assertNil(t, err)
	assertEqual(t, want, string(res))
}

func TestMarshalerSlice(t *testing.T) {
	s := testStructMarshaler{ID: 100}
	s1 := testStructMarshaler{ID: 101}

	r := []testStructMarshaler{s, s1}

	want := `{"id":"100","type":"test-structs","attributes":{"name":"changed"}}`
	want1 := `{"id":"101","type":"test-structs","attributes":{"name":"changed"}}`
	want = "[" + want + "," + want1 + "]"
	res, err := MarshalSlice(&r)
	assertNil(t, err)
	assertEqual(t, want, string(res))
}

func TestMarshal(t *testing.T) {
	s := testStruct{
		ID:       100,
		Name:     "John",
		Address:  testSub{Country: "CTR", City: "DT"},
		Excluded: "dont include this",
	}

	want := `{"id":"100","type":"test-structs","attributes":{"name":"John","address-at":{"country":"CTR","city":"DT"}}}`

	res, err := Marshal(&s)
	assertNil(t, err)
	assertEqual(t, want, string(res))
}

func TestMarshalNonAPI(t *testing.T) {
	s := testStructNonAPI{
		ID:       100,
		Name:     "John",
		Address:  testSub{Country: "CTR", City: "DT"},
		Excluded: "include this",
	}

	want := `{"ID":100,"Name":"John","Address":{"country":"CTR","city":"DT"},"Excluded":"include this"}` + "\n"

	res, err := Marshal(&s)
	assertNil(t, err)
	assertEqual(t, want, string(res))
}

func TestMarshalError(t *testing.T) {
	_, err := Marshal(nil)
	assertEqual(t, errMarshalInvalidData, err)
}

func TestMarshalSlice(t *testing.T) {
	s := testStruct{
		ID:       100,
		Name:     "John",
		Address:  testSub{Country: "CTR", City: "DT"},
		Excluded: "dont include this",
	}
	s1 := testStruct{
		ID:       101,
		Name:     "John1",
		Address:  testSub{Country: "CTR1", City: "DT1"},
		Excluded: "dont include this",
	}

	r := []testStruct{s, s1}
	want := `{"id":"100","type":"test-structs","attributes":{"name":"John","address-at":{"country":"CTR","city":"DT"}}}`
	want1 := `{"id":"101","type":"test-structs","attributes":{"name":"John1","address-at":{"country":"CTR1","city":"DT1"}}}`

	want = "[" + want + "," + want1 + "]"

	res, err := MarshalSlice(r)
	assertNil(t, err)
	assertEqual(t, want, string(res))
}

func TestMarshalSliceError(t *testing.T) {
	_, err := MarshalSlice(nil)
	assertEqual(t, errMarshalInvalidData, err)
	_, err = MarshalSlice("asd")
	assertEqual(t, errMarshalInvalidData, err)
	_, err = MarshalSlice([]string{"asd"})
	assertEqual(t, errMarshalInvalidData, err)
}

func BenchmarkMarshal(b *testing.B) {
	s := &testStruct{
		ID:       100,
		Name:     "John",
		Address:  testSub{Country: "CTR", City: "DT"},
		Excluded: "dont include this",
	}
	for i := 0; i < b.N; i++ {
		Marshal(s)
	}
}
