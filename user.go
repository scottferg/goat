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
