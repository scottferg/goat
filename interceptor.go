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
	"bytes"
	"encoding/base64"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strings"
)

type Interceptor func(http.ResponseWriter, *http.Request, *Context) Handler

func Generic401(w http.ResponseWriter, r *http.Request, c *Context) error {
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	return nil
}

func Generic403(w http.ResponseWriter, r *http.Request, c *Context) error {
	http.Error(w, "Forbidden", http.StatusForbidden)
	return nil
}

func NewBasicAuthInterceptor(normal Handler) Interceptor {
	return func(w http.ResponseWriter, r *http.Request, c *Context) Handler {
		auth := r.Header.Get("Authorization")

		if auth == "" {
			return Generic403
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
			return Generic401
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
				if c.Database.C("goat_users").Find(bson.M{"_id": uid}).One(&c.User); c.User != nil {
					return normal
				}
			}
		}

		return unauthorized
	}
}
