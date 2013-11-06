package logger

import (
	"fmt"
	"io"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strings"
)

type DBLogger struct {
	database   *mgo.Database
	collection string
}

type Entry struct {
	Id      bson.ObjectId `bson:"_id"`
	Message string        `bson:"message"`
}

func New(c string, db *mgo.Database) *DBLogger {
	return &DBLogger{
		database:   db,
		collection: c,
	}
}

func (l *DBLogger) Write(data []byte) (int, error) {
	err := l.database.C(l.collection).Insert(Entry{
		Id:      bson.NewObjectId(),
		Message: strings.TrimSpace(string(data)),
	})

	if err != nil {
		fmt.Println(err.Error())
	}
	return len(data), err
}

func Tail(out io.Writer, log string, db *mgo.Database) {
	var entry Entry
	// Find the last entry in the tailable collection, then
	// use that to determine where to begin the cursor
	db.C(log).Find(nil).Sort("-$natural").Limit(1).One(&entry)

	query := func(id string) *mgo.Query {
		return db.C(log).Find(bson.M{
			"_id": bson.M{
				"$gt": id,
			},
		})
	}

	iter := query(entry.Message).Sort("$natural").Tail(-1)
	for {
		for iter.Next(&entry) {
			fmt.Fprintf(out, entry.Message)

			if err := iter.Close(); err != nil {
				fmt.Println(err)
				return
			}

			iter = query(entry.Message).Sort("$natural").Tail(-1)
		}
	}
}
