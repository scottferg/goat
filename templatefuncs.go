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

func ParseTemplates(name, path string) *template.Template {
    t := template.New(name).Funcs(funcMap)
    t, err := t.ParseGlob(path)

    return template.Must(t, err)
}

func ObjectIdHex(id bson.ObjectId) string {
	return id.Hex()
}
