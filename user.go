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
	"code.google.com/p/go.crypto/bcrypt"
	"labix.org/v2/mgo/bson"
	"net/http"
)

type User struct {
	Id       bson.ObjectId `bson:"_id,omitempty"`
	Username string
	Password []byte
}

// SetPassword takes a plaintext password and hashes it with bcrypt and sets the
// password field to the hash.
func (u *User) SetPassword(password string) error {
	hpass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = hpass

	return nil
}

func (u *User) Save(c *Context) error {
	if err := c.Database.C("goat_users").Insert(u); err != nil {
		return err
	}

	return nil
}

func (u *User) Login(w http.ResponseWriter, r *http.Request, c *Context) {
    c.Session.Values["uid"] = u.Id
    c.Session.Save(r, w)
}

func NewUser(username, password string) (u *User, err error) {
	u = &User{
		Id:       bson.NewObjectId(),
		Username: username,
	}
	err = u.SetPassword(password)

	return
}

// Login validates and returns a user object if they exist in the database.
func Authenticate(c *Context, username, password string) (u *User, err error) {
	if err = c.Database.C("goat_users").Find(bson.M{"username": username}).One(&u); err != nil {
		return
	}

	if err = bcrypt.CompareHashAndPassword(u.Password, []byte(password)); err != nil {
		u = nil
	}

	return
}
