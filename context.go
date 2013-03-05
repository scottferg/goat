package goat 

import (
	"github.com/gorilla/sessions"
	"labix.org/v2/mgo"
	"net/http"
)

type Context struct {
	Database *mgo.Database
	Session  *sessions.Session
	User     *User
}

func (c *Context) Close() {
    if c.Database != nil {
        c.Database.Session.Close()
    }
}

func NewContext(req *http.Request) (*Context, error) {
	return new(Context), nil
}
