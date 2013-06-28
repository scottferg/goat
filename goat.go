/****************************************************************************
 * Copyright (c) 2013, Scott Ferguson
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *     * Redistributions of source code must retain the above copyright
 *       notice, this list of conditions and the following disclaimer.
 *     * Redistributions in binary form must reproduce the above copyright
 *       notice, this list of conditions and the following disclaimer in the
 *       documentation and/or other materials provided with the distribution.
 *     * Neither the name of the software nor the
 *       names of its contributors may be used to endorse or promote products
 *       derived from this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY SCOTT FERGUSON ''AS IS'' AND ANY
 * EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
 * WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL SCOTT FERGUSON BE LIABLE FOR ANY
 * DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
 * (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
 * LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
 * ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
 * SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 ****************************************************************************/

package goat

import (
	"encoding/gob"
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/scottferg/mux"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

const (
	GET    = 0x1
	POST   = 0x2
	PUT    = 0x4
	DELETE = 0x8

	methodGet    = "GET"
	methodPost   = "POST"
	methodPut    = "PUT"
	methodDelete = "DELETE"
)

type Goat struct {
	Router       *mux.Router
	routes       map[string]*route
	middleware   []Middleware
	dbsession    *mgo.Session
	dbname       string
	sessionstore sessions.Store
	listener     *net.TCPListener
}

type Handler func(http.ResponseWriter, *http.Request, *Context) error

type route struct {
	*Goat
	path        string
	name        string
	handler     Handler
	interceptor Interceptor
}

func (r route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var c *Context
	var err error

	if c, err = NewContext(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer c.Close()

	// Execute Middleware
	for _, m := range r.middleware {
		m(req, c)
	}

	// Execute the handler
	if r.handler != nil {
		err = r.handler(w, req, c)
	} else if r.interceptor != nil {
		rh := r.interceptor(w, req, c)
		err = rh(w, req, c)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func methodList(methods int) (r []string) {
	if methods&GET == GET {
		r = append(r, methodGet)
	}

	if methods&POST == POST {
		r = append(r, methodPost)
	}

	if methods&PUT == PUT {
		r = append(r, methodPut)
	}

	if methods&DELETE == DELETE {
		r = append(r, methodDelete)
	}

	return
}

func NewGoat() *Goat {
	// Initialize session store
	gob.Register(bson.ObjectId(""))
	s := sessions.NewCookieStore([]byte("sevenbyelevensecretbomberboy"))
	r := mux.NewRouter()

	http.Handle("/", r)

	return &Goat{
		Router:       r,
		sessionstore: s,
		routes:       make(map[string]*route),
	}
}

func (g *Goat) RegisterRoute(path, name string, method int, handler interface{}) {
	// Initialize the HTTP router
	r := new(route)
	r.Goat = g
	r.path = path
	r.name = name

	if g.routes[r.name] != nil {
		return
	}

	g.routes[r.name] = r

	if h, ok := handler.(func(http.ResponseWriter, *http.Request, *Context) error); ok {
		r.handler = h
	} else if h, ok := handler.(Handler); ok {
		r.handler = h
	} else if i, ok := handler.(Interceptor); ok {
		r.interceptor = i
	} else if h, ok := handler.(http.Handler); ok {
		g.Router.Handle(path, h)
		return
	} else {
		panic("Unknown handler passed to RegisterRoute")
	}

	methods := methodList(method)
	g.Router.Handle(path, r).Methods(methods...).Name(r.name)
}

func (g *Goat) CloneDB() *mgo.Database {
	return g.dbsession.Clone().DB(g.dbname)
}

func (g *Goat) RegisterStaticFileHandler(remote, local string) {
	// Static file handler
	http.Handle(remote, http.FileServer(http.Dir(local)))
}

func (g *Goat) RegisterMiddleware(m Middleware) {
	g.middleware = append(g.middleware, m)
}

func (g *Goat) Reverse(root string, params ...string) (*url.URL, error) {
	return g.Router.Get(root).URL(params...)
}

func (g *Goat) ListenAndServe(port string) {
	p := 8080

	if port != "" {
		p, _ = strconv.Atoi(port)
	}

	server := &http.Server{}

	g.listener, _ = net.ListenTCP("tcp", &net.TCPAddr{
		Port: p,
	})

	if err := server.Serve(g.listener); err != nil {
		fmt.Printf("Error when starting server: %s", err.Error())
	}
}

func (g *Goat) Close() {
	g.listener.Close()
}
