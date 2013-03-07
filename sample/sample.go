package main

import (
	"errors"
	"fmt"
	"github.com/vokalinteractive/goat"
	"net/http"
)

var (
    g *goat.Goat
)

func Index(w http.ResponseWriter, r *http.Request, c *goat.Context) error {
    fmt.Fprint(w, "You got here to the index!")

    return nil
}

func ErrorRoute(w http.ResponseWriter, r *http.Request, c *goat.Context) error {
    return errors.New("This is a 500! Goat handles your errors for you! Neat-o!")
}

func main() {
	g = goat.NewGoat()

	g.RegisterMiddleware(g.NewSessionMiddleware("my-goat-session"))

	g.RegisterRoute("/error", "error", goat.GET, ErrorRoute)
	g.RegisterRoute("/", "index", goat.GET, Index)

	g.ListenAndServe("8080")
}
