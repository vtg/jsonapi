package jsonapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
)

var (
	errMarshalInvalidData = errors.New("jsonapi: invalid data structure passed for marshalling")
)

// Marshal item to json api format
func Marshal(i interface{}) ([]byte, error) {
	v := interfacePtr(i)
	if !v.IsValid() {
		return []byte{}, errMarshalInvalidData
	}

	c := encoder{buf: new(bytes.Buffer)}
	if err := c.marshal(v); err != nil {
		return []byte{}, err
	}

	return c.buf.Bytes(), nil
}

// MarshalSlice marshalling items to json api format
func MarshalSlice(i interface{}) ([]byte, error) {
	e := interfacePtr(i)

	if !e.IsValid() {
		return []byte{}, errMarshalInvalidData
	}

	if e.Type().Kind() == reflect.Ptr {
		e = e.Elem()
	}

	if e.Type().Kind() != reflect.Slice {
		return []byte{}, errMarshalInvalidData
	}

	c := encoder{buf: new(bytes.Buffer)}
	c.buf.WriteByte('[')
	iLen := e.Len()
	for i := 0; i < iLen; i++ {
		if err := c.marshal(valuePtr(e.Index(i))); err != nil {
			return []byte{}, err
		}
		if i < iLen-1 {
			c.buf.WriteByte(',')
		}
	}
	c.buf.WriteByte(']')
	return c.buf.Bytes(), nil
}

type encoder struct {
	buf *bytes.Buffer
}

func (e *encoder) marshal(el reflect.Value) error {
	t := el.Type()
	if t.Implements(beforeMarshalerType) {
		m := el.Interface().(BeforeMarshaler)
		if err := m.BeforeMarshalJSONAPI(); err != nil {
			return err
		}
	}
	if t.Implements(marshalerType) {
		m := el.Interface().(Marshaler)
		b, err := m.MarshalJSONAPI()
		if err != nil {
			return err
		}
		e.buf.Write(b)
		return nil
	}

	if t.Kind() == reflect.Ptr {
		el = el.Elem()
	}

	t = el.Type()
	if t.Kind() != reflect.Struct {
		return errMarshalInvalidData
	}

	f := types.get(t)
	if !f.api() {
		return json.NewEncoder(e.buf).Encode(el.Interface())
	}

	e.buf.WriteByte('{')
	e.buf.WriteString(`"id":"`)

	id := el.FieldByIndex(f.id)
	switch id.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		e.buf.WriteString(strconv.FormatUint(id.Uint(), 10))
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		e.buf.WriteString(strconv.FormatInt(id.Int(), 10))
	case reflect.String:
		e.buf.WriteString(id.String())
	}
	e.buf.WriteString(`","type":"`)
	e.buf.WriteString(f.stype)
	aLen := len(f.attrs)
	if aLen > 0 {
		e.buf.WriteString(`","attributes":{`)
		for k := range f.attrs {
			e.buf.WriteByte('"')
			e.buf.WriteString(f.attrs[k].name)
			e.buf.WriteByte('"')
			e.buf.WriteByte(':')
			b, err := json.Marshal(el.FieldByIndex(f.attrs[k].idx).Interface())
			if err != nil {
				return err
			}
			e.buf.Write(b)
			if k < aLen-1 {
				e.buf.WriteByte(',')
			}
		}
		e.buf.WriteByte('}')
	}
	e.buf.WriteByte('}')

	return nil
}
