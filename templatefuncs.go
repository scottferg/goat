package goat

import (
	"html/template"
	"labix.org/v2/mgo/bson"
)

var (
	funcMap = template.FuncMap{
		"objectIdHex": ObjectIdHex,
	}
)

func ParseTemplates(path string) *template.Template {
    t := template.New("templates").Funcs(funcMap)
    t, err := t.ParseGlob("templates/*")

    return template.Must(t, err)
}

func ObjectIdHex(id bson.ObjectId) string {
	return id.Hex()
}
