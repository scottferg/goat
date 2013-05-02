/****************************************************************************
 * Copyright (c) 2013, VOKAL Interactive, LLC
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
 * THIS SOFTWARE IS PROVIDED BY VOKAL INTERACTIVE, LLC ''AS IS'' AND ANY
 * EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
 * WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL VOKAL INTERACTIVE, LLC BE LIABLE FOR ANY
 * DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
 * (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
 * LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
 * ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
 * SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 ****************************************************************************/

package goat

import (
	"labix.org/v2/mgo"
	"net/http"
	"net/url"
)

type Middleware func(*http.Request, *Context) error

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

	if name == "" {
		parsed, err := url.Parse(host)

		if err == nil {
			name = parsed.Path[1:]
		} else {
			name = host
		}
	}

	if err != nil {
		panic(err.Error())
	}

	return func(r *http.Request, c *Context) error {
		c.Database = g.dbsession.Clone().DB(name)
		return nil
	}
}
