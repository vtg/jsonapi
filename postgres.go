package jsonapi

import (
	"bytes"
	"encoding/json"
	"reflect"
	"unicode"
)

// PGRelations type to store jsonapi relations
type PGRelations map[string]interface{}

// PostgresJSON returns postgres prepared json object
func PostgresJSON(i interface{}, prefix string, relations PGRelations) (string, error) {
	v := interfacePtr(i)
	if !v.IsValid() {
		return "", errMarshalInvalidData
	}
	c := &encoder{}
	if err := c.sql(v, prefix, relations); err != nil {
		return "", err
	}

	return c.String(), nil
}

func (e *encoder) sql(el reflect.Value, prefix string, relations PGRelations) error {
	t := el.Type()

	if t.Kind() == reflect.Ptr {
		el = el.Elem()
	}

	t = el.Type()
	if t.Kind() != reflect.Struct {
		return errMarshalInvalidData
	}

	f := types.get(t)
	if !f.api() {
		b, err := json.Marshal(el.Interface())
		e.Write(b)
		return err
	}

	e.WriteString("json_build_object(")
	e.WriteString("'id',")
	e.WriteString(prefix)
	e.WriteString(f.idName)
	e.WriteString("::TEXT")
	e.WriteByte(',')
	e.WriteString("'type',")
	e.WriteByte('\'')
	e.WriteString(f.stype)
	e.WriteByte('\'')

	aLen := len(f.attrs)
	if aLen > 0 {
		e.WriteByte(',')
		e.WriteString("'attributes',json_build_object(")
		for k := range f.attrs {
			if f.attrs[k].create {
				continue
			}
			ev := el.FieldByIndex(f.attrs[k].idx)
			if f.attrs[k].skipEmpty && isEmptyValue(ev) {
				continue
			}
			e.WriteByte('\'')
			e.WriteString(f.attrs[k].name)
			e.WriteByte('\'')
			e.WriteByte(',')
			if !f.attrs[k].skipPrefix {
				e.WriteString(prefix)
			}
			e.WriteString(f.attrs[k].dbName)
			if f.attrs[k].quote {
				e.WriteString("::TEXT")
			}
			if k < aLen-1 {
				e.WriteByte(',')
			}
		}
		e.WriteByte(')')
	}
	aLen = len(relations)
	if aLen > 0 {
		idx := 0
		e.WriteString(`,'relationships',json_build_object(`)
		for k, v := range relations {
			el := interfacePtr(v)
			t := el.Type()

			if t.Kind() == reflect.Ptr {
				el = el.Elem()
			}

			t = el.Type()
			switch t.Kind() {
			case reflect.Struct:
				f := types.get(t)
				e.WriteByte('\'')
				e.WriteString(f.stype)
				e.WriteByte('\'')
				e.WriteString(",json_build_object('data',array_to_json(array_agg(")
				en := encoder{}
				en.sql(el, k, PGRelations{})
				e.Write(en.Bytes())
				e.WriteByte(')')
				e.WriteByte(')')
				e.WriteByte(')')
			case reflect.String:
				e.WriteByte('\'')
				e.WriteString(k)
				e.WriteByte('\'')
				e.WriteString(",json_build_object('links',json_build_object('related',")
				e.WriteString(v.(string))
				e.WriteByte(')')
				e.WriteByte(')')
			}

			if idx < aLen-1 {
				e.WriteByte(',')
			}
			idx++
		}
		e.WriteByte(')')

	}

	e.WriteByte(')')

	return nil
}

func columnName(s string) string {
	var buf bytes.Buffer
	var idx byte
	var r rune
	for i, v := range s {
		if (i > 1 && idx == 0x1) || (idx == 0x2 && unicode.IsLower(v)) {
			buf.WriteByte('_')
		}
		if i > 0 {
			buf.WriteRune(r)
		}
		if unicode.IsUpper(v) {
			if idx == 0x0 {
				idx = 0x1
			} else {
				idx = 0x2
			}
			r = unicode.ToLower(v)
		} else {
			idx = 0x0
			r = v
		}
	}
	buf.WriteRune(r)
	return buf.String()
}
