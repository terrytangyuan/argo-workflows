package main

import (
	"log"

	"github.com/upper/db/v4/adapter/postgresql"
)

var settings = postgresql.ConnectionURL{
	Database: "testdb",
	Host:     "127.0.0.1",
	User:     "akp-tndb001-15-tst-usw2-app@akuity-test.iam",
	Password: "Go3AdiNoK9fhD@nByae@!XmZDJqUFP",
	Options:  map[string]string{"sslmode": "disable", "port": "1234"},
}

func main() {
	sess, err := postgresql.Open(settings)
	if err != nil {
		log.Fatal("postgresql.Open: ", err)
	}
	defer sess.Close()
	err = sess.Ping()
	if err != nil {
		log.Fatal("ping: ", err)
	}
}
