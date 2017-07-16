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
	ID   uint64   `jsonapi:"id,test-rels"`
	Name string   `jsonapi:"attr,name"`
	Rel1 Relation `jsonapi:"rel,rel1"`
	// Rel2 Links `jsonapi:"rellink,rel2"`
}

func TestMarshalRelations(t *testing.T) {
	s := testRelations{
		ID:   100,
		Name: "A",
		// Rel2: Links{Self: "self/2",Related: "rel/2"},
	}

	want := `{"id":"100","type":"test-rels","attributes":{"name":"A"},"relationships":{"rel1":{}}}`
	res, err := Marshal(&s)
	assertNil(t, err)
	assertEqual(t, want, string(res))

	s.Rel1.Links.Self = "self/1"
	s.Rel1.Links.Related = "rel/1"

	want = `{"id":"100","type":"test-rels","attributes":{"name":"A"},"relationships":{"rel1":{"links":{"self":"self/1","related":"rel/1"}}}}`
	res, err = Marshal(&s)
	assertNil(t, err)
	assertEqual(t, want, string(res))

	s.Rel1.Data = "reldata"

	want = `{"id":"100","type":"test-rels","attributes":{"name":"A"},"relationships":{"rel1":{"links":{"self":"self/1","related":"rel/1"},"data":"reldata"}}}`
	res, err = Marshal(&s)
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

func TestMarshalWithScope(t *testing.T) {
	s := struct {
		ID   uint64 `jsonapi:"id,test-structs"`
		S1   string `jsonapi:"attr,s1"`
		S2   string `jsonapi:"attr,s2" scope:"2"`
		S3   string `jsonapi:"attr,s3" scope:"3"`
		Both string `jsonapi:"attr,both" scope:"2,3"`
	}{
		ID:   100,
		S1:   "v1",
		S2:   "v2",
		S3:   "v3",
		Both: "v4",
	}

	want := `{"id":"100","type":"test-structs","attributes":{"s1":"v1","s2":"v2","s3":"v3","both":"v4"}}`
	res, err := MarshalWithScope(&s, "")
	assertNil(t, err)
	assertEqual(t, want, string(res))

	want = `{"id":"100","type":"test-structs","attributes":{"s1":"v1","s2":"v2","both":"v4"}}`
	res, err = MarshalWithScope(&s, "2")
	assertNil(t, err)
	assertEqual(t, want, string(res))

	want = `{"id":"100","type":"test-structs","attributes":{"s1":"v1","s3":"v3","both":"v4"}}`
	res, err = MarshalWithScope(&s, "3")
	assertNil(t, err)
	assertEqual(t, want, string(res))

	want = `{"id":"100","type":"test-structs","attributes":{"s1":"v1"}}`
	res, err = MarshalWithScope(&s, "4")
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
