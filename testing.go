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
	"fmt"
	"github.com/scottferg/mux"
	"net/http"
	"net/http/httptest"
	"time"
)

type TestSuite struct {
	g *Goat
}

func NewTestSuite(g *Goat) *TestSuite {
	return &TestSuite{
		g,
	}
}

func (t *TestSuite) SetUp() {
	fmt.Println("Goat setup!")
	t.g.RegisterMiddleware(t.g.NewDatabaseMiddleware("localhost", fmt.Sprintf("%s_goat_test", time.Now().String())))
}

func (t *TestSuite) TearDown() {
	fmt.Println("Goat teardown!")
}

func (g *Goat) ServeRoute(name string, vars map[string]string) (*httptest.Server, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	route := g.routes[name]

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if vars != nil {
			mux.SetVars(r, vars)
		}

		route.ServeHTTP(recorder, r)
	})), recorder
}
