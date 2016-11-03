package jsonapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testPGType struct {
	ID     uint64 `jsonapi:"id,types1"`
	Name   string `jsonapi:"attr,name"`
	APIAge int    `jsonapi:"attr,age,string"`
}

type testPGType2 struct {
	ID   uint64 `jsonapi:"id,types2"`
	Name string `jsonapi:"attr,name"`
}

func TestPostgresObject(t *testing.T) {
	tp := testPGType{
		ID:     100,
		Name:   "John",
		APIAge: 100,
	}
	want := `json_build_object('id',t1.id::TEXT,'type','types1','attributes',json_build_object('name',t1.name,'age',t1.api_age::TEXT))`
	got, err := PostgresJSON(&tp, "t1.", JSONStruct{})
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestPostgresObjectExtraAttrs(t *testing.T) {
	tp := testPGType{
		ID:     100,
		Name:   "John",
		APIAge: 100,
	}
	want := `json_build_object('id',t1.id::TEXT,'type','types1','attributes',json_build_object('name',t1.name,'age',t1.api_age::TEXT,'extra',col))`
	got, err := PostgresJSON(&tp, "t1.", JSONStruct{
		Attributes: map[string]string{"extra": "col"},
	})
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestPostgresObjectWithRelations(t *testing.T) {
	tp := testPGType{
		ID:     100,
		Name:   "John",
		APIAge: 100,
	}
	want := `json_build_object('id',t1.id::TEXT,'type','types1','attributes',json_build_object('name',t1.name,'age',t1.api_age::TEXT),'relationships',json_build_object('types2',json_build_object('data',array_to_json(array_agg(json_build_object('id',t4.id::TEXT,'type','types2','attributes',json_build_object('name',t4.name)))))))`
	got, err := PostgresJSON(&tp, "t1.", JSONStruct{
		Relations: map[string]interface{}{
			"t4.": &testPGType2{},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestPostgresObjectWithRelationLinks(t *testing.T) {
	tp := testPGType{
		ID:     100,
		Name:   "John",
		APIAge: 100,
	}
	want := `json_build_object('id',t1.id::TEXT,'type','types1','attributes',json_build_object('name',t1.name,'age',t1.api_age::TEXT),'relationships',json_build_object('types2',json_build_object('links',json_build_object('related','string/asdf'))))`
	got, err := PostgresJSON(&tp, "t1.", JSONStruct{
		Relations: map[string]interface{}{
			"types2": "'string/asdf'",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestColumnName(t *testing.T) {
	ts := [][]string{
		[]string{"Id", "id"},
		[]string{"ID", "id"},
		[]string{"Name", "name"},
		[]string{"NameId", "name_id"},
		[]string{"NameID", "name_id"},
		[]string{"JSON", "json"},
		[]string{"SomeJSON", "some_json"},
		[]string{"CamelCase", "camel_case"},
		[]string{"SomeLongJSONStringFormat", "some_long_json_string_format"},
		[]string{"Some123Number", "some123_number"},
	}

	for _, v := range ts {
		got := columnName(v[0])
		assert.Equal(t, v[1], got)
	}
}

func BenchmarkColumnName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		columnName("SomeLongJSONStringFormat")
	}
}
