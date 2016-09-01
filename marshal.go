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

	c := &encoder{}
	if err := c.marshal(v); err != nil {
		return []byte{}, err
	}

	return c.Bytes(), nil
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

	c := &encoder{}
	c.WriteByte('[')
	iLen := e.Len()
	for i := 0; i < iLen; i++ {
		if err := c.marshal(valuePtr(e.Index(i))); err != nil {
			return []byte{}, err
		}
		if i < iLen-1 {
			c.WriteByte(',')
		}
	}
	c.WriteByte(']')
	return c.Bytes(), nil
}

type encoder struct {
	bytes.Buffer
	buffer [64]byte
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
		e.Write(b)
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
		b, err := json.Marshal(el.Interface())
		e.Write(b)
		return err
	}

	e.WriteByte('{')
	e.WriteString(`"id":"`)

	id := el.FieldByIndex(f.id)
	switch id.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		e.Write(strconv.AppendUint(e.buffer[:0], id.Uint(), 10))
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		e.Write(strconv.AppendInt(e.buffer[:0], id.Int(), 10))
	case reflect.String:
		e.WriteString(id.String())
	}
	e.WriteString(`","type":"`)
	e.WriteString(f.stype)
	aLen := len(f.attrs)
	if aLen > 0 {
		e.WriteString(`","attributes":{`)
		for k := range f.attrs {
			e.WriteByte('"')
			e.WriteString(f.attrs[k].name)
			e.WriteByte('"')
			e.WriteByte(':')
			b, err := json.Marshal(el.FieldByIndex(f.attrs[k].idx).Interface())
			if err != nil {
				return err
			}
			if f.attrs[k].quote {
				e.WriteByte('"')
			}
			e.Write(b)
			if f.attrs[k].quote {
				e.WriteByte('"')
			}
			if k < aLen-1 {
				e.WriteByte(',')
			}
		}
		e.WriteByte('}')
	}
	aLen = len(f.links)
	if aLen > 0 {
		e.WriteString(`,"links":{`)
		for k := range f.links {
			e.WriteByte('"')
			e.WriteString(f.links[k].name)
			e.WriteByte('"')
			e.WriteByte(':')
			b, err := json.Marshal(el.FieldByIndex(f.links[k].idx).Interface())
			if err != nil {
				return err
			}
			e.Write(b)
			if k < aLen-1 {
				e.WriteByte(',')
			}
		}
		e.WriteByte('}')
	}
	aLen = len(f.rels)
	if aLen > 0 {
		e.WriteString(`,"relationships":{`)
		for k := range f.rels {
			e.WriteByte('"')
			e.WriteString(f.rels[k].name)
			e.WriteByte('"')
			e.WriteByte(':')
			if f.rels[k].link {
				e.WriteString(`{"links":`)
			}
			b, err := json.Marshal(el.FieldByIndex(f.rels[k].idx).Interface())
			if err != nil {
				return err
			}
			e.Write(b)
			if f.rels[k].link {
				e.WriteByte('}')

			}
			if k < aLen-1 {
				e.WriteByte(',')
			}
		}
		e.WriteByte('}')
	}
	e.WriteByte('}')

	return nil
}
