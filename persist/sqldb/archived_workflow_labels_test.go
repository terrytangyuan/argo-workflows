package sqldb

import (
	"cloud.google.com/go/cloudsqlconn"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"

	"context"
	"k8s.io/apimachinery/pkg/labels"
)

func Test_labelsClause(t *testing.T) {
	// Ref: https://cloud.google.com/sql/docs/postgres/iam-logins#log-in-with-automatic
	d, err := cloudsqlconn.NewDialer(context.Background(), cloudsqlconn.WithIAMAuthN())
	assert.Nil(t, err)
	var opts []cloudsqlconn.DialOption

	dsn := fmt.Sprintf("user=%s database=%s", "p774999980047-mame9g@gcp-sa-cloud-sql.iam.gserviceaccount.com", "postgres")
	instanceConnectionName := "akuity-test:us-west2:akp-tenantdb001-15-tst-usw2-primary"
	config, err := pgx.ParseConfig(dsn)
	assert.Nil(t, err)

	config.DialFunc = func(ctx context.Context, network, instance string) (net.Conn, error) {
		return d.Dial(ctx, instanceConnectionName, opts...)
	}
	println(config.ConnString())
	dbURI := stdlib.RegisterConnConfig(config)
	// "missing \"=\" after \"registeredConnConfig0\" in connection info string\"
	//connURL, err := postgresql.ParseURL(dbURI)
	//assert.Nil(t, err)
	//options := map[string]string{
	//	"sslmode": "disable",
	//}
	//connURL.Options = options
	//_, err = postgresql.Open(connURL)
	db, err := sql.Open("pgx", dbURI)
	assert.Nil(t, err)
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
