package jsonapi

import (
	"reflect"
	"strings"
	"sync"
	"unicode"
)

var types = typesCache{m: make(map[reflect.Type]*fields)}

// Marshaler interface
type Marshaler interface {
	MarshalJSONAPI() ([]byte, error)
}

// BeforeMarshaler interface
type BeforeMarshaler interface {
	BeforeMarshalJSONAPI() error
}

// Unmarshaler interface
type Unmarshaler interface {
	UnmarshalJSONAPI([]byte) error
}

// BeforeUnmarshaler interface
type BeforeUnmarshaler interface {
	BeforeUnmarshalJSONAPI() error
}

var (
	marshalerType         = reflect.TypeOf(new(Marshaler)).Elem()
	beforeMarshalerType   = reflect.TypeOf(new(BeforeMarshaler)).Elem()
	unmarshalerType       = reflect.TypeOf(new(Unmarshaler)).Elem()
	beforeUnmarshalerType = reflect.TypeOf(new(BeforeUnmarshaler)).Elem()
)

// MetaData struct
type MetaData struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// Response structure for json api response
type Response struct {
	Data   interface{} `json:"data,omitempty"`
	Errors interface{} `json:"errors,omitempty"`
	Meta   *MetaData   `json:"meta,omitempty"`
}

type fields struct {
	id    []int
	stype string
	attrs []field
}

func (f fields) api() bool {
	return len(f.id) > 0
}

type field struct {
	idx      []int
	name     string
	readonly bool
}

type typesCache struct {
	sync.RWMutex
	m map[reflect.Type]*fields
}

func (s *typesCache) get(t reflect.Type) *fields {
	s.RLock()
	f := types.m[t]
	s.RUnlock()

	if f != nil {
		return f
	}

	s.Lock()

	f = &fields{}

	for _, idx := range typeFields(t, []int{}) {
		fd := t.FieldByIndex(idx)
		tag := fd.Tag.Get("jsonapi")
		if tag == "" {
			continue
		}

		keys := strings.SplitN(tag, ",", 3)
		switch keys[0] {
		case "id":
			f.id = idx
			if len(keys) > 1 {
				f.stype = keys[1]
			}
		case "attr":
			name := fd.Name
			if len(keys) > 1 && validKey(keys[1]) {
				name = keys[1]
			}
			ro := false
			if len(keys) > 2 {
				ro = keys[2] == "readonly"
			}
			f.attrs = append(f.attrs, field{idx: idx, name: name, readonly: ro})
		}
	}
	s.m[t] = f

	s.Unlock()

	return f
}

func validKey(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		switch {
		case strings.ContainsRune("!#$%&()*+-./:<=>?@[]^_{|}~ ", c):
		default:
			if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
				return false
			}
		}
	}
	return true
}

func interfacePtr(i interface{}) reflect.Value {
	v := reflect.ValueOf(i)
	if !v.IsValid() {
		return v
	}
	return valuePtr(v)
}

func valuePtr(v reflect.Value) reflect.Value {
	if v.Type().Kind() != reflect.Ptr && v.CanAddr() {
		return v.Addr()
	}
	return v
}

func typeFields(t reflect.Type, idx []int) (res [][]int) {
	for i := 0; i < t.NumField(); i++ {
		fd := t.Field(i)

		if fd.PkgPath != "" && !fd.Anonymous {
			continue
		}

		idx1 := append(idx, fd.Index...)
		if fd.Anonymous && fd.Type.Kind() == reflect.Struct {
			for _, v := range typeFields(fd.Type, idx1) {
				res = append(res, v)
			}
			continue
		}

		res = append(res, idx1)
	}
	return
}
