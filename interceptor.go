package goat

import (
	"bytes"
	"encoding/base64"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strings"
)

type Interceptor func(http.ResponseWriter, *http.Request, *Context) Handler

func NewBasicAuthInterceptor(normal Handler) Interceptor {
	unauthorized := func(w http.ResponseWriter, r *http.Request, c *Context) error {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil
	}

	return func(w http.ResponseWriter, r *http.Request, c *Context) Handler {
		auth := r.Header.Get("Authorization")

		if auth == "" {
			return unauthorized
		}

		encoded := strings.Split(auth, "Basic ")[1]
		reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(encoded))

		buf := new(bytes.Buffer)
		buf.ReadFrom(reader)
		s := buf.String()

		credentials := strings.Split(s, ":")

		// Authenticate now
		u, err := Authenticate(c, credentials[0], credentials[1])
		if err != nil {
			return unauthorized
		}

		c.User = u
		return normal
	}
}

func NewAuthSessionInterceptor(normal, unauthorized Handler) Interceptor {
	return func(w http.ResponseWriter, r *http.Request, c *Context) Handler {
		// Check if a user has been authenticated, otherwise
		// redirect to the unauthorized view
		uid := c.Session.Values["uid"]
		if uid != nil {
			// run the handler and grab the error, and report it
			if uid, ok := c.Session.Values["uid"].(bson.ObjectId); ok {
                // TODO: Error check here
				c.Database.C("goat_users").Find(bson.M{"_id": uid}).One(&c.User)
			}

			return normal
		}

		return unauthorized
	}
}
