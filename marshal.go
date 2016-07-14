package jsonapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
)

// Marshal item to json api format
func Marshal(i interface{}) ([]byte, error) {
	v := ptrValue(i)
	if !v.IsValid() {
		return []byte{}, errors.New("jsonapi: only struct allowed for parsing")
	}

	e := v.Elem()

	return marshal(e)
}

// MarshalSlice marshalling items to json api format
func MarshalSlice(i interface{}) ([]byte, error) {
	e := ptrValue(i)

	if !e.IsValid() || e.Type().Kind() != reflect.Slice {
		return []byte{}, errors.New("jsonapi: only slice allowed for parsing")
	}

	var buf bytes.Buffer
	buf.WriteByte('[')
	iLen := e.Len()
	for i := 0; i < iLen; i++ {
		b, err := marshal(e.Index(i))
		if err != nil {
			return []byte{}, err
		}
		buf.Write(b)
		if i < iLen-1 {
			buf.WriteByte(',')
		}
	}
	buf.WriteByte(']')
	return buf.Bytes(), nil
}

func marshal(e reflect.Value) ([]byte, error) {
	t := e.Type()
	if t.Kind() != reflect.Struct {
		return []byte{}, errors.New("jsonapi: only struct allowed for parsing")
	}

	f := types.get(t)
	if !f.api() {
		return json.Marshal(e.Interface())
	}

	var buf bytes.Buffer
	buf.WriteByte('{')
	buf.WriteString(`"id":"`)

	id := e.FieldByIndex(f.id)
	switch id.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		buf.WriteString(strconv.FormatUint(id.Uint(), 10))
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		buf.WriteString(strconv.FormatInt(id.Int(), 10))
	case reflect.String:
		buf.WriteString(id.String())
	}
	buf.WriteString(`","type":"`)
	buf.WriteString(f.stype)
	aLen := len(f.attrs)
	if aLen > 0 {
		buf.WriteString(`","attributes":{`)
		for k := range f.attrs {
			buf.WriteByte('"')
			buf.WriteString(f.attrs[k].name)
			buf.WriteByte('"')
			buf.WriteByte(':')
			b, err := json.Marshal(e.FieldByIndex(f.attrs[k].idx).Interface())
			if err != nil {
				return []byte{}, err
			}
			buf.Write(b)
			if k < aLen-1 {
				buf.WriteByte(',')
			}
		}
		buf.WriteByte('}')
	}
	buf.WriteByte('}')

	return buf.Bytes(), nil
}
