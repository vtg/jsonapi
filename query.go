package jsonapi

import (
	"strconv"
	"strings"
)

// Keymap type for storing key-value
type Keymap [2]string

// Key returns key
func (k Keymap) Key() string {
	return k[0]
}

// Value returns value
func (k Keymap) Value() string {
	return k[1]
}

// Query contains parsed url query params
type Query struct {
	Limit   int
	Offset  int
	Format  string
	Sort    []string
	Filters []Keymap
	Queries []Keymap
}

// DefaultSort set default sort column
func (q *Query) DefaultSort(s string) {
	if s != "" && len(q.Sort) == 0 {
		q.Sort = []string{s}
	}
}

// DefaultLimit set default limit
func (q *Query) DefaultLimit(n int) {
	if q.Limit == 0 {
		q.Limit = 30
	}
}

// QueryParams takes URL params and returns parsed params for jsonapi
func QueryParams(m map[string][]string) *Query {
	p := new(Query)
	p.Sort = make([]string, 0, 2)
	p.Queries = make([]Keymap, 0, 2)
	p.Filters = make([]Keymap, 0, 2)

	for key, params := range m {
		ln := len(params[0])
		switch key {
		case "sort":
			idx := 0
			for i := 0; i < ln; i++ {
				if params[0][i] == ',' || i == ln-1 {
					var w string
					if i == ln-1 {
						w = params[0][idx:]
					} else {
						w = params[0][idx:i]
					}
					if w[0] == '-' {
						p.Sort = append(p.Sort, w[1:]+" DESC")
					} else {
						p.Sort = append(p.Sort, w)
					}
					idx = i + 1
				}
			}
		case "limit":
			p.Limit = strToInt(params[0])
		case "offset":
			p.Offset = strToInt(params[0])
		case "format":
			p.Format = params[0]
		default:
			// filtering
			if strings.HasPrefix(key, "filter[") {
				val := strings.TrimSuffix(strings.TrimPrefix(key, "filter["), "]")
				p.Filters = append(p.Filters, Keymap{val, params[0]})
				continue
			}
			// search queries
			if strings.HasPrefix(key, "query[") {
				val := strings.TrimSuffix(strings.TrimPrefix(key, "query["), "]")
				p.Queries = append(p.Queries, Keymap{val, params[0]})
			}
		}
	}
	return p
}

func strToInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
