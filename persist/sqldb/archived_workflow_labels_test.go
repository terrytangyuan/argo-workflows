package sqldb

import (
	"cloud.google.com/go/cloudsqlconn"
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/stretchr/testify/assert"
	postgresqladp "github.com/upper/db/v4/adapter/postgresql"
	"k8s.io/apimachinery/pkg/labels"
	"net"
	"testing"
	"upper.io/db.v3/postgresql"
)

func Test_labelsClause(t *testing.T) {
	assert.Equal(t, 1, 1)
	// Ref: https://cloud.google.com/sql/docs/postgres/iam-logins#log-in-with-automatic
	d, err := cloudsqlconn.NewDialer(context.Background(), cloudsqlconn.WithIAMAuthN())
	assert.Nil(t, err)
	var opts []cloudsqlconn.DialOption

	dsn := "user=akp-tndb001-15-tst-usw2-app@akuity-test.iam database=testdb sslmode=disable"
	instanceConnectionName := "akuity-test:us-west2:akp-tenantdb001-15-tst-usw2"
	config, err := pgx.ParseConfig(dsn)
	assert.Nil(t, err)

	config.DialFunc = func(ctx context.Context, network, instance string) (net.Conn, error) {
		return d.Dial(ctx, instanceConnectionName, opts...)
	}
	println(config.ConnString())
	dbURI := stdlib.RegisterConnConfig(config)
	// "missing \"=\" after \"registeredConnConfig0\" in connection info string\"
	connURL, err := postgresqladp.ParseURL(dbURI)
	assert.Nil(t, err)
	options := map[string]string{
		"sslmode": "disable",
	}
	connURL.Options = options
	db, err := postgresqladp.Open(
		postgresql.ConnectionURL{
			Database: `testdb`,
			User:     `akp-tndb001-15-tst-usw2-app@akuity-test.iam`,
			Options:  options,
		})
	//db, err := sql.Open("pgx", dbURI)
	//assert.Nil(t, err)
	// failed SASL auth
	err = db.Ping()
	assert.Nil(t, err)
}

func requirements(selector string) []labels.Requirement {
	requirements, err := labels.ParseToRequirements(selector)
	if err != nil {
		panic(err)
	}
	return requirements
}
