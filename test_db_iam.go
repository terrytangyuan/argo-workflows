package main

import (
	"cloud.google.com/go/cloudsqlconn"
	"context"
	"database/sql"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"log"
	"net"
)

func main() {
	// Ref: https://cloud.google.com/sql/docs/postgres/iam-logins#log-in-with-automatic
	d, err := cloudsqlconn.NewDialer(context.Background(), cloudsqlconn.WithIAMAuthN())
	if err != nil {
		log.Fatal("NewDialer: ", err)
	}
	var opts []cloudsqlconn.DialOption
	dsn := "host=34.94.245.198 user=akp-tndb001-15-tst-usw2-app@akuity-test.iam database=testdb sslmode=disable"
	instanceConnectionName := "akuity-test:us-west2:akp-tenantdb001-15-tst-usw2"
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		log.Fatal("ParseConfig: ", err)
	}
	config.DialFunc = func(ctx context.Context, network, instance string) (net.Conn, error) {
		return d.Dial(ctx, instanceConnectionName, opts...)
	}
	dbURI := stdlib.RegisterConnConfig(config)
	sess, err := sql.Open("pgx", dbURI)
	if err != nil {
		log.Fatal("sql.Open: ", err)
	}
	defer sess.Close()
	err = sess.Ping()
	if err != nil {
		log.Fatal("ping: ", err)
	}
}
