package jsonapi

import "testing"

type testSub struct {
	Country string `json:"country"`
	City    string `json:"city"`
}

type testStruct struct {
	ID        uint64  `jsonapi:"id,test-structs"`
	Name      string  `jsonapi:"attr,name"`
	Address   testSub `jsonapi:"attr,address-at"`
	IntString uint64  `jsonapi:"attr,intstring,readonly,string"`
	Excluded  string
}

type testStructOmit struct {
	ID      uint64   `jsonapi:"id,test-omits"`
	Name    string   `jsonapi:"attr,name,omitempty"`
	Address *testSub `jsonapi:"attr,address,omitempty"`
	Int     int      `jsonapi:"attr,int,omitempty"`
	Uint64  uint64   `jsonapi:"attr,uint64,omitempty"`
	Uint32  uint32   `jsonapi:"attr,uint32,omitempty"`
}

type testLinkStruct struct {
	ID   uint64 `jsonapi:"id,test-structs"`
	Name string `jsonapi:"attr,name"`
	Self string `jsonapi:"link,self"`
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

func (t *testStructMarshaler) MarshalJSONAPI() ([]byte, error) {
	return []byte("custom"), nil
}

type testStructBeforeMarshaler struct {
	ID   uint64 `jsonapi:"id,test-structs"`
	Name string `jsonapi:"attr,name"`
}

func (t *testStructBeforeMarshaler) BeforeMarshalJSONAPI() error {
	t.Name = "changed"
	return nil
}

func TestMarshaler(t *testing.T) {
	s := testStructMarshaler{ID: 100}
	want := `custom`
	res, err := Marshal(&s)
	assertNil(t, err)
	assertEqual(t, want, string(res))
}

func TestBeforeMarshaler(t *testing.T) {
	s := testStructBeforeMarshaler{ID: 100}
	want := `{"id":"100","type":"test-structs","attributes":{"name":"changed"}}`
	res, err := Marshal(&s)
	assertNil(t, err)
	assertEqual(t, want, string(res))
}

type testRelations struct {
	ID   uint64 `jsonapi:"id,test-rels"`
	Name string `jsonapi:"attr,name"`
	Rel1 Links  `jsonapi:"rellink,rel1"`
	// Rel2 Links `jsonapi:"rellink,rel2"`
}

func TestMarshalRelations(t *testing.T) {
	s := testRelations{
		ID:   100,
		Name: "A",
		Rel1: Links{Self: "self/1", Related: "rel/1"},
		// Rel2: Links{Self: "self/2",Related: "rel/2"},
	}
	want := `{"id":"100","type":"test-rels","attributes":{"name":"A"},"relationships":{"rel1":{"links":{"self":"self/1","related":"rel/1"}}}}`
	res, err := Marshal(&s)
	assertNil(t, err)
	assertEqual(t, want, string(res))
}

func TestMarshalerSlice(t *testing.T) {
	s := testStructBeforeMarshaler{ID: 100}
	s1 := testStructBeforeMarshaler{ID: 101}

	r := []testStructBeforeMarshaler{s, s1}

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

	want := `{"id":"100","type":"test-structs","attributes":{"name":"John","address-at":{"country":"CTR","city":"DT"},"intstring":"0"}}`

	res, err := Marshal(&s)
	assertNil(t, err)
	assertEqual(t, want, string(res))
}

func TestMarshalOmit(t *testing.T) {
	s := testStructOmit{
		ID: 100,
	}

	want := `{"id":"100","type":"test-omits","attributes":{}}`

	res, err := Marshal(&s)
	assertNil(t, err)
	assertEqual(t, want, string(res))
}

func TestMarshalLinks(t *testing.T) {
	s := testLinkStruct{
		ID:   100,
		Name: "John",
		Self: "/test/1",
	}

	want := `{"id":"100","type":"test-structs","attributes":{"name":"John"},"links":{"self":"/test/1"}}`

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

	want := `{"ID":100,"Name":"John","Address":{"country":"CTR","city":"DT"},"Excluded":"include this"}`

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
	want := `{"id":"100","type":"test-structs","attributes":{"name":"John","address-at":{"country":"CTR","city":"DT"},"intstring":"0"}}`
	want1 := `{"id":"101","type":"test-structs","attributes":{"name":"John1","address-at":{"country":"CTR1","city":"DT1"},"intstring":"0"}}`

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
