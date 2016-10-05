package jsonapi

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"sync"
	"unicode"
)

var types = typesCache{m: make(map[reflect.Type]*fields)}

// Marshaler interface example
// 	func (p *Post) MarshalJSONAPI() ([]byte, error) {
// 		return []byte(`{"custom":"return"}`), nil
// 	}
type Marshaler interface {
	MarshalJSONAPI() ([]byte, error)
}

// BeforeMarshaler interface
// 	func (p *Post) BeforeMarshalJSONAPI() error {
// 		p.SelfLink = fmt.Sprintf("/api/posts/%d", p.ID)
// 		p.Comments.Related = fmt.Sprintf("/api/posts/%d/comments", p.ID)
// 		return nil
// 	}
type BeforeMarshaler interface {
	BeforeMarshalJSONAPI() error
}

// Unmarshaler interface
type Unmarshaler interface {
	UnmarshalJSONAPI([]byte) error
}

// AfterUnmarshaler interface
type AfterUnmarshaler interface {
	AfterUnmarshalJSONAPI() error
}

var (
	marshalerType        = reflect.TypeOf(new(Marshaler)).Elem()
	beforeMarshalerType  = reflect.TypeOf(new(BeforeMarshaler)).Elem()
	unmarshalerType      = reflect.TypeOf(new(Unmarshaler)).Elem()
	afterUnmarshalerType = reflect.TypeOf(new(AfterUnmarshaler)).Elem()
)

// MetaData struct
type MetaData struct {
	Total  int         `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
	Data   interface{} `json:"data,omitempty"`
}

// Response structure for json api response
type Response struct {
	Data     interface{} `json:"data,omitempty"`
	Included interface{} `json:"included,omitempty"`
	Meta     *MetaData   `json:"meta,omitempty"`
	Errors
}

// StatusCode returns first error status code or success
func (r Errors) StatusCode() int {
	if r.HasErrors() {
		return strToInt(r.Errors[0].Status)
	}
	return 200
}

// Links structure
type Links struct {
	Self    string `json:"self,omitempty"`
	Related string `json:"related,omitempty"`
}

// Relation structure
type Relation struct {
	Links Links
	Data  interface{}
}

// MarshalJSON marshaller
func (r Relation) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	if r.Links.Self != "" || r.Links.Related != "" {
		buf.WriteString(`"links":{`)
		if r.Links.Self != "" {
			buf.WriteString(`"self":"`)
			buf.WriteString(r.Links.Self)
			buf.WriteByte('"')
			if r.Links.Related != "" {
				buf.WriteByte(',')
			}
		}
		if r.Links.Related != "" {
			buf.WriteString(`"related":"`)
			buf.WriteString(r.Links.Related)
			buf.WriteByte('"')
		}
		buf.WriteByte('}')
		if r.Data != nil {
			buf.WriteByte(',')
		}
	}
	if r.Data != nil {
		buf.WriteString(`"data":`)
		b, err := json.Marshal(r.Data)
		if err != nil {
			return []byte{}, err
		}
		buf.Write(b)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

type fields struct {
	id    []int
	stype string
	attrs []field
	links []field
	rels  []field
}

func (f fields) api() bool {
	return len(f.id) > 0
}

type field struct {
	idx       []int
	name      string
	readonly  bool
	quote     bool
	link      bool
	skipEmpty bool
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

		keys := strings.Split(tag, ",")
		switch keys[0] {
		case "id":
			f.id = idx
			if len(keys) > 1 {
				f.stype = keys[1]
			}
		case "attr":
			fld := field{idx: idx, name: fd.Name}
			if len(keys) > 1 && validKey(keys[1]) {
				fld.name = keys[1]
			}
			if len(keys) > 2 {
				for _, v := range keys[2:] {
					switch v {
					case "readonly":
						fld.readonly = true
					case "string":
						fld.quote = true
					case "omitempty":
						fld.skipEmpty = true
					}
				}
			}
			f.attrs = append(f.attrs, fld)
		case "link":
			name := fd.Name
			if len(keys) > 1 && validKey(keys[1]) {
				name = keys[1]
			}
			f.links = append(f.links, field{idx: idx, name: name})
		case "rel":
			name := fd.Name
			if len(keys) > 1 && validKey(keys[1]) {
				name = keys[1]
			}
			f.rels = append(f.rels, field{idx: idx, name: name})
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
