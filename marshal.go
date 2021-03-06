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

// MarshalWithScope item to json api format
func MarshalWithScope(i interface{}, scope string) ([]byte, error) {
	return marshalWithScope(i, scope)
}

// Marshal item to json api format
func Marshal(i interface{}) ([]byte, error) {
	return marshalWithScope(i, "")
}

// Marshal item to json api format
func marshalWithScope(i interface{}, scope string) ([]byte, error) {
	e := interfacePtr(i)
	if !e.IsValid() {
		return []byte{}, errMarshalInvalidData
	}

	e1 := e
	if e.Type().Kind() == reflect.Ptr {
		e1 = e.Elem()
	}

	c := &encoder{}
	switch e1.Type().Kind() {
	case reflect.Slice, reflect.Array:
		c.WriteByte('[')
		iLen := e1.Len()
		for i := 0; i < iLen; i++ {
			if err := c.marshal(valuePtr(e1.Index(i)), scope); err != nil {
				return []byte{}, err
			}
			if i < iLen-1 {
				c.WriteByte(',')
			}
		}
		c.WriteByte(']')
	default:
		if err := c.marshal(e, scope); err != nil {
			return []byte{}, err
		}
	}

	return c.Bytes(), nil
}

type encoder struct {
	bytes.Buffer
	buffer [64]byte
}

func (e *encoder) marshal(el reflect.Value, scope string) error {
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

	f := types.get(el)
	if !f.api() {
		b, err := json.Marshal(el.Interface())
		e.Write(b)
		return err
	}

	e.WriteByte('{')
	e.WriteString(`"id":`)

	id := el.FieldByIndex(f.id)
	if id.Type().Implements(jsonMarshallerType) {
		m := id.Interface().(json.Marshaler)
		b, _ := m.MarshalJSON()
		e.Write(b)
	} else {
		e.WriteByte('"')
		switch id.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			e.Write(strconv.AppendUint(e.buffer[:0], id.Uint(), 10))
		case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
			e.Write(strconv.AppendInt(e.buffer[:0], id.Int(), 10))
		case reflect.String:
			e.WriteString(id.String())
		}
		e.WriteByte('"')
	}
	e.WriteString(`,"type":"`)
	e.WriteString(f.stype)
	if len(f.attrs) > 0 {
		empty := true
		e.WriteString(`","attributes":{`)
		for k := range f.attrs {
			ev := el.FieldByIndex(f.attrs[k].idx)
			if f.attrs[k].skipEmpty && isEmptyValue(ev) {
				continue
			}
			if !f.attrs[k].inScope(scope) {
				continue
			}
			if !empty {
				e.WriteByte(',')
			}
			e.WriteByte('"')
			e.WriteString(f.attrs[k].name)
			e.WriteByte('"')
			e.WriteByte(':')
			b, err := json.Marshal(ev.Interface())
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
			empty = false
		}
		e.WriteByte('}')
	}
	if len(f.links) > 0 {
		e.WriteString(`,"links":{`)
		for k := range f.links {
			if k > 0 {
				e.WriteByte(',')
			}
			e.WriteByte('"')
			e.WriteString(f.links[k].name)
			e.WriteByte('"')
			e.WriteByte(':')
			b, err := json.Marshal(el.FieldByIndex(f.links[k].idx).Interface())
			if err != nil {
				return err
			}
			e.Write(b)
		}
		e.WriteByte('}')
	}
	if len(f.rels) > 0 {
		e.WriteString(`,"relationships":{`)
		for k := range f.rels {
			if k > 0 {
				e.WriteByte(',')
			}
			e.WriteByte('"')
			e.WriteString(f.rels[k].name)
			e.WriteByte('"')
			e.WriteByte(':')
			b, err := json.Marshal(el.FieldByIndex(f.rels[k].idx).Interface())
			if err != nil {
				return err
			}
			e.Write(b)
		}
		e.WriteByte('}')
	}
	e.WriteByte('}')

	return nil
}

// isEmptyValue taken from go standard encoding/json package
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
