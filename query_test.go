package jsonapi

import (
	"net/http"
	"strings"
	"testing"
)

func httpRequest(method, url string, body string) *http.Request {
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		panic(err)
	}
	return req
}

func TestKeymapBlank(t *testing.T) {
	k := Keymap{}
	assertEqual(t, "", k.Key())
	assertEqual(t, "", k.Value())
	k = Keymap{"key", "val"}
	assertEqual(t, "key", k.Key())
	assertEqual(t, "val", k.Value())
}

func TestURLParams(t *testing.T) {
	req := httpRequest("GET", "/page?format=short&query[name]=john&query[email]=john1&filter[active]=1&filter[id]=1,2,3&limit=10&offset=1&sort=-name,id", "")
	p := QueryParams(req.URL.Query())
	assertEqual(t, "short", p.Format)
	assertEqual(t, 10, p.Limit)
	assertEqual(t, 1, p.Offset)
	assertEqual(t, []string{"name DESC", "id"}, p.Sort)
	assertEqual(t, []Keymap{Keymap{"active", "1"}, Keymap{"id", "1,2,3"}}, p.Filters)
	assertEqual(t, []Keymap{Keymap{"name", "john"}, Keymap{"email", "john1"}}, p.Queries)
}

func BenchmarkQueryParams(b *testing.B) {
	req := httpRequest("GET", "/page?format=short&query[name]=john&query[email]=john1&filter[active]=1&filter[id]=1,2,3&limit=10&offset=1&sort=-name,id", "")
	q := req.URL.Query()
	for i := 0; i < b.N; i++ {
		QueryParams(q)
	}
}
