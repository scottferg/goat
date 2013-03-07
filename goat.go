package goat

import (
	"encoding/gob"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"net/url"
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
	router       *mux.Router
	middleware   []Middleware
	dbsession    *mgo.Session
	sessionstore sessions.Store
}

type Handler func(http.ResponseWriter, *http.Request, *Context) error

type Route struct {
	*Goat
	handler     Handler
	interceptor Interceptor
}

func (r Route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var c *Context
	var err error

	if c, err = NewContext(req); err != nil {
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

func getMethodList(methods int) (r []string) {
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
		router:       r,
		sessionstore: s,
	}
}

func (g *Goat) RegisterRoute(path, name string, method int, handler interface{}) {
	// Initialize the HTTP router
	r := new(Route)
	r.Goat = g

	if h, ok := handler.(func(http.ResponseWriter, *http.Request, *Context) error); ok {
		r.handler = h
	} else if h, ok := handler.(Handler); ok {
		r.handler = h
	} else if i, ok := handler.(Interceptor); ok {
		r.interceptor = i
	}

	methods := getMethodList(method)
	g.router.Handle(path, r).Methods(methods...)
}

func (g *Goat) RegisterStaticFileHandler(path string) {
	// Static file handler
	http.Handle(path, http.StripPrefix(path, http.FileServer(http.Dir("."+path))))
}

func (g *Goat) RegisterMiddleware(m Middleware) {
	g.middleware = append(g.middleware, m)
}

func (g *Goat) Reverse(route string) (*url.URL, error) {
	return g.router.Get("index").URL()
}

func (g *Goat) ListenAndServe(port string) {
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr: ":" + port,
	}

	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("Error when starting server: %s", err.Error())
	}
}
