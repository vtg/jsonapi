package jsonapi

import (
	"reflect"
	"strings"
	"sync"
	"unicode"
)

var types = typesCache{m: make(map[reflect.Type]*fields)}

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
	idx  []int
	name string
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

	for i := 0; i < t.NumField(); i++ {
		fd := t.Field(i)

		if fd.PkgPath != "" && !fd.Anonymous {
			continue
		}

		tag := fd.Tag.Get("jsonapi")
		if tag == "" {
			continue
		}

		keys := strings.SplitN(tag, ",", 2)
		switch keys[0] {
		case "id":
			f.id = fd.Index
			if len(keys) > 1 {
				f.stype = keys[1]
			}
		case "attr":
			name := fd.Name
			if len(keys) > 1 && validKey(keys[1]) {
				name = keys[1]
			}
			f.attrs = append(f.attrs, field{idx: fd.Index, name: name})
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

func ptrValue(i interface{}) reflect.Value {
	v := reflect.ValueOf(i)
	if !v.IsValid() {
		return v
	}
	if v.Type().Kind() != reflect.Ptr {
		if v.CanAddr() {
			return v.Addr()
		}
	}
	return v
}
