package jsonapi

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

// Request structure for unmarshaling
type Request struct {
	Data struct {
		ID         json.Number                `json:"id"`
		Type       string                     `json:"type"`
		Attributes map[string]json.RawMessage `json:"attributes"`
	} `json:"data"`
}

// Change structure for storing structure changes
type Change struct {
	Field string
	Cur   string
	New   string
}

func (c Change) equal() bool {
	return c.Cur == c.New
}

// // String method
// func (c Change) String() string {
// 	return c.Field + ": " + c.Cur + " -> " + c.New
// }

// UnmarshalWithScope decoding json api compatible request
func UnmarshalWithScope(b []byte, i interface{}, scope string) error {
	return unmarshal(b, i, scope)
}

// Unmarshal decoding json api compatible request
func Unmarshal(b []byte, i interface{}) error {
	return unmarshal(b, i, "")
}

// unmarshal decoding json api compatible request
func unmarshal(b []byte, i interface{}, scope string) error {
	v := interfacePtr(i)
	if !v.IsValid() {
		return errMarshalInvalidData
	}

	d := decoder{}
	return d.unmarshal(b, v, scope)
}

// UnmarshalWithChangesWithScope decoding json api compatible request into structure
// and returning changes
func UnmarshalWithChangesWithScope(b []byte, i interface{}, scope string) (Changes, error) {
	return unmarshalWithChanges(b, i, scope)
}

// UnmarshalWithChanges decoding json api compatible request into structure
// and returning changes
func UnmarshalWithChanges(b []byte, i interface{}) (Changes, error) {
	return unmarshalWithChanges(b, i, "")
}

func unmarshalWithChanges(b []byte, i interface{}, scope string) (Changes, error) {
	v := interfacePtr(i)
	if !v.IsValid() {
		return Changes{}, errMarshalInvalidData
	}
	d := decoder{withChanges: true}
	err := d.unmarshal(b, v, scope)
	return d.changes, err
}

// Changes store
type Changes []Change

// Empty returns true if there are no changes
func (c Changes) Empty() bool {
	return len(c) == 0
}

// Find changed field by name
func (c Changes) Find(k string) Change {
	for i := range c {
		if c[i].Field == k {
			return c[i]
		}
	}
	return Change{}
}

// Contains returns true if changes have field
func (c Changes) Contains(keys ...string) bool {
	for _, k := range keys {
		for i := range c {
			if c[i].Field == k {
				return true
			}
		}
	}
	return false
}

// // String method
// func (c Changes) String() string {
// 	strs := make([]string, 0, len(c))
// 	for i := range c {
// 		strs = append(strs, c[i].String())
// 	}
// 	return strings.Join(strs, "\n")
// }

type decoder struct {
	withChanges bool
	changes     Changes
}

// Unmarshal decoding json api compatible request
func (d *decoder) unmarshal(b []byte, e reflect.Value, scope string) error {
	t := e.Type()

	if t.Implements(unmarshalerType) {
		m := e.Interface().(Unmarshaler)
		return m.UnmarshalJSONAPI(b)
	}

	e1 := e
	if t.Kind() == reflect.Ptr {
		e1 = e.Elem()
	}

	t1 := e1.Type()
	if t1.Kind() != reflect.Struct {
		return errMarshalInvalidData
	}

	f := types.get(e1)
	if !f.api() {
		return fmt.Errorf("jsonapi: %v incompatible with json api", t1.Name())
	}

	req := Request{}
	err := json.Unmarshal(b, &req)
	if err != nil {
		return err
	}

	if req.Data.Type != f.stype {
		return fmt.Errorf("jsonapi: can't unmarshal item of type '%s' into item of type '%s'", req.Data.Type, f.stype)
	}

	ne := reflect.New(t1).Elem()

	if d.withChanges {
		d.changes = make([]Change, 0, len(f.attrs))
	}

	for _, attr := range f.attrs {
		if !attr.readonly {
			v, ok := req.Data.Attributes[attr.name]
			if !ok {
				continue
			}

			if !attr.inScope(scope) {
				continue
			}

			curVal := e1.FieldByIndex(attr.idx)
			newVal := ne.FieldByIndex(attr.idx)
			if attr.quote {
				v = unquote(v)
			}
			err = json.Unmarshal(v, newVal.Addr().Interface())
			if err != nil {
				return err
			}

			if d.withChanges {
				d.diff(curVal, newVal, attr.name)
			}

			curVal.Set(newVal)
		}
	}

	if e.Type().Implements(afterUnmarshalerType) {
		return e.Interface().(AfterUnmarshaler).AfterUnmarshalJSONAPI()
	}

	return nil
}

func unquote(b []byte) []byte {
	l := len(b)
	if l > 1 && b[0] == '"' && b[l-1] == '"' {
		return b[1 : l-1]
	}
	return b
}

type mapKeys struct {
	keys []reflect.Value
}

func (m *mapKeys) add(v reflect.Value) {
	for k := range m.keys {
		if m.keys[k].String() == v.String() {
			return
		}
	}

	m.keys = append(m.keys, v)
}

func boolString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func (d *decoder) diff(v1, v2 reflect.Value, field string) bool {
	v1 = getElement(v1)
	v2 = getElement(v2)
	change := Change{Field: field}

	switch v1.Type().Kind() {
	case reflect.Map:
		keys := mapKeys{}
		if v1.Kind() == v2.Kind() {
			for _, v := range v1.MapKeys() {
				keys.add(v)
			}
			for _, v := range v2.MapKeys() {
				keys.add(v)
			}
		}
		for _, key := range keys.keys {
			d.diff(v1.MapIndex(key), v2.MapIndex(key), change.Field+"."+key.String())
		}
	case reflect.Struct:
		if v1.Type().Name() == "Time" {
			change.Cur = stringVal(v1)
			change.New = stringVal(v2)
		} else {
			t := v1.Type()
			for i := 0; i < t.NumField(); i++ {
				fd := t.Field(i)
				// skip ignored fields
				if tag := fd.Tag.Get("bson"); tag == "-" {
					continue
				}

				if (fd.PkgPath != "" && !fd.Anonymous) || v1.Kind() != v2.Kind() {
					continue
				}
				d.diff(v1.FieldByIndex(fd.Index), v2.FieldByIndex(fd.Index), change.Field+"."+fd.Name)
			}
		}
	default:
		change.Cur = stringVal(v1)
		change.New = stringVal(v2)
	}

	if !change.equal() {
		d.changes = append(d.changes, change)
	}
	return !change.equal()
}

func getElement(v reflect.Value) reflect.Value {
	if !v.IsValid() {
		return reflect.ValueOf("")
	}

	k := v.Type().Kind()
	if k == reflect.Ptr || k == reflect.Interface {
		return getElement(v.Elem())
	}

	return v
}

func stringVal(v reflect.Value) string {
	if v.Type().Implements(stringerType) {
		return v.Interface().(stringer).String()
	}

	switch v.Type().Kind() {
	case reflect.String:
		return v.String()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Bool:
		return boolString(v.Bool())
	}
	return fmt.Sprintf("%v", v.Interface())
}
