/*
Package jsonapi is simple JSON API implementation for GO. To read more about JSON API format please visit http://jsonapi.org/

Example:

  import (
  	"fmt"

  	"github.com/vtg/jsonapi"
  )

  type Post struct {
  	ID       uint64        `jsonapi:"id,users"`
  	Name     string        `jsonapi:"attr,name"`
  	Age      uint32        `jsonapi:"attr,age,string,readonly"` // this will be marshalled as string and will be ignored on unmarshal
  	SelfLink string        `jsonapi:"link,self"`
  	Comments jsonapi.Links `jsonapi:"rellink,comments"`
  }

  // BeforeMarshalJSONAPI will be executed before marshalling
  func (p *Post) BeforeMarshalJSONAPI() error {
  	p.SelfLink = fmt.Sprintf("/api/posts/%d", p.ID)
  	p.Comments.Related = fmt.Sprintf("/api/posts/%d/comments", p.ID)
  	return nil
  }

  // AfterUnmarshalJSONAPI will be executed after unmarshalling
  func (p *Post) AfterUnmarshalJSONAPI() error {
  	v := jsonapi.Validator{}
  	v.Present(p.Name, "name")
  	return v.Verify()
  }


*/
package jsonapi
