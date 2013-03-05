package goat

import (
	"labix.org/v2/mgo"
	"net/http"
	"net/url"
)

func (g *Goat) NewSessionMiddleware(storename string) Middleware {
	return func(r *http.Request, c *Context) error {
		s, err := g.sessionstore.Get(r, storename)
		c.Session = s

		return err
	}
}

func (g *Goat) NewDatabaseMiddleware(host, name string) Middleware {
	// When registering we'll establish the database
	// connection. We'll clone it on each request.
	s, err := mgo.Dial(host)
	if err != nil {
		panic(err.Error())
	}

	g.dbsession = s

	return func(r *http.Request, c *Context) error {
		var err error
		if name == "" {
			parsed, err := url.Parse(host)

			if err == nil {
				name = parsed.Path[1:]
			} else {
				name = host
			}
		}

		c.Database = g.dbsession.Clone().DB(name)

		return err
	}
}
