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
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"labix.org/v2/mgo/bson"
	"net/http"
	"time"
)

type ResetToken struct {
	Id        bson.ObjectId
	Username  string
	Token     string
	Timestamp time.Time
}

func (r *ResetToken) Delete(c *Context) error {
	return c.Database.C("goat_reset_tokens").RemoveId(r.Id)
}

func (r *ResetToken) Save(c *Context) error {
	_, err := c.Database.C("goat_reset_tokens").UpsertId(r.Id, r)
	return err
}

type User struct {
	Id       bson.ObjectId          `json:"-" bson:"_id,omitempty"`
	Username string                 `json:"username,omitempty" bson:"username,omitempty"`
	Password []byte                 `json:"-" bson:"password,omitempty"`
	Values   map[string]interface{} `json:"values,omitempty" bson:"values,omitempty"`
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

func (u *User) Save(c *Context) (err error) {
	_, err = c.Database.C("goat_users").UpsertId(u.Id, u)

	return
}

func (u *User) Login(w http.ResponseWriter, r *http.Request, c *Context) {
	c.Session.Values["uid"] = u.Id
	c.Session.Save(r, w)
}

func NewUser(username, password string, c *Context) (u *User, err error) {
	query := c.Database.C("goat_users").Find(bson.M{"username": bson.M{"$in": []string{username}}})
	if n, _ := query.Count(); n > 0 {
		return nil, errors.New("account with that name already exists")
	}

	u = &User{
		Id:       bson.NewObjectId(),
		Username: username,
		Values:   make(map[string]interface{}),
	}
	err = u.SetPassword(password)

	return
}

func FindUser(username string, c *Context) (u *User, err error) {
	err = c.Database.C("goat_users").Find(bson.M{
		"username": username,
	}).One(&u)

	return
}

// Login validates and returns a user object if they exist in the database.
func Authenticate(username, password string, c *Context) (u *User, err error) {
	if err = c.Database.C("goat_users").Find(bson.M{"username": username}).One(&u); err != nil {
		return
	}

	if err = bcrypt.CompareHashAndPassword(u.Password, []byte(password)); err != nil {
		u = nil
	}

	return
}

func ResetPassword(token, password string, c *Context) error {
	// Get the reset token record
	var reset ResetToken
	err := c.Database.C("goat_reset_tokens").Find(
		bson.M{
			"token": token,
		}).One(&reset)
	if err != nil {
		return err
	}

	// Token expiry is 48 hours
	if time.Since(reset.Timestamp) > (48 * time.Hour) {
		reset.Delete(c)
		return errors.New("token expired")
	}

	// Get the user
	u, err := FindUser(reset.Username, c)
	if err != nil {
		return err
	}

	err = u.SetPassword(password)
	if err != nil {
		return err
	}

	err = reset.Delete(c)
	if err != nil {
		return err
	}

	return u.Save(c)
}

// Fetches a request token for the user. If the user is found,
// they will be added to the provided context.
func RequestResetToken(username string, c *Context) (*ResetToken, error) {
	u, err := FindUser(username, c)
	if err != nil {
		return nil, err
	}
	c.User = u

	token := ResetToken{
		Id:        bson.NewObjectId(),
		Username:  u.Username,
		Timestamp: time.Now(),
	}

	hash := md5.New()
	io.WriteString(hash, token.Id.Hex())
	io.WriteString(hash, token.Timestamp.String())

	token.Token = fmt.Sprintf("%x", hash.Sum(nil))

	token.Save(c)

	return &token, nil
}
