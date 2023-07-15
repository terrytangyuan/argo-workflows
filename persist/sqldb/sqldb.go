package sqldb

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"k8s.io/client-go/kubernetes"
	"net"
	"time"
	"upper.io/db.v3/lib/sqlbuilder"
	"upper.io/db.v3/mysql"
	"upper.io/db.v3/postgresql"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/argoproj/argo-workflows/v3/config"
	"github.com/argoproj/argo-workflows/v3/errors"
	"github.com/argoproj/argo-workflows/v3/util"
)

func GetTableName(persistConfig *config.PersistConfig) (string, error) {
	var tableName string
	if persistConfig.PostgreSQL != nil {
		tableName = persistConfig.PostgreSQL.TableName

	} else if persistConfig.MySQL != nil {
		tableName = persistConfig.MySQL.TableName
	}
	if tableName == "" {
		return "", errors.InternalError("TableName is empty")
	} else {
		return tableName, nil
	}
}

// CreateDBSession creates the dB session
func CreateDBSession(kubectlConfig kubernetes.Interface, namespace string, persistConfig *config.PersistConfig) (sqlbuilder.Database, error) {
	if persistConfig == nil {
		return nil, errors.InternalError("Persistence config is not found")
	}

	if persistConfig.PostgreSQL != nil {
		return CreatePostGresDBSession(kubectlConfig, namespace, persistConfig.PostgreSQL, persistConfig.ConnectionPool)
	} else if persistConfig.MySQL != nil {
		return CreateMySQLDBSession(kubectlConfig, namespace, persistConfig.MySQL, persistConfig.ConnectionPool)
	}
	return nil, fmt.Errorf("no databases are configured")
}

// CreatePostGresDBSession creates postgresDB session
func CreatePostGresDBSession(kubectlConfig kubernetes.Interface, namespace string, cfg *config.PostgreSQLConfig, persistPool *config.ConnectionPool) (sqlbuilder.Database, error) {
	ctx := context.Background()
	userNameByte, err := util.GetSecrets(ctx, kubectlConfig, namespace, cfg.UsernameSecret.Name, cfg.UsernameSecret.Key)
	if err != nil {
		return nil, err
	}
	passwordByte, err := util.GetSecrets(ctx, kubectlConfig, namespace, cfg.PasswordSecret.Name, cfg.PasswordSecret.Key)
	if err != nil {
		return nil, err
	}

	// Ref: https://cloud.google.com/sql/docs/postgres/iam-logins#log-in-with-automatic
	d, err := cloudsqlconn.NewDialer(context.Background(), cloudsqlconn.WithIAMAuthN())
	if err != nil {
		return nil, fmt.Errorf("cloudsqlconn.NewDialer: %w", err)
	}
	var opts []cloudsqlconn.DialOption

	dsn := fmt.Sprintf("user=%s database=%s", "service-account-name@project-id.iam", "dbname")
	instanceConnectionName := "project:region:instance"
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	config.DialFunc = func(ctx context.Context, network, instance string) (net.Conn, error) {
		return d.Dial(ctx, instanceConnectionName, opts...)
	}
	dbURI := stdlib.RegisterConnConfig(config)
	connURL, err := postgresql.ParseURL(dbURI)
	if err != nil {
		return nil, fmt.Errorf("postgresql.ParseURL: %w", err)
	}
	session, err := postgresql.Open(connURL)
	if err != nil {
		return nil, fmt.Errorf("postgresql.Open: %w", err)
	}
	return session, nil

	settings := postgresql.ConnectionURL{
		User:     string(userNameByte),
		Password: string(passwordByte),
		Host:     cfg.GetHostname(),
		Database: cfg.Database,
	}

	if cfg.SSL {
		if cfg.SSLMode != "" {
			options := map[string]string{
				"sslmode": cfg.SSLMode,
			}
			settings.Options = options
		}
	}

	session, err := postgresql.Open(settings)
	if err != nil {
		return nil, err
	}
	session = ConfigureDBSession(session, persistPool)
	return session, nil
}

// CreateMySQLDBSession creates Mysql DB session
func CreateMySQLDBSession(kubectlConfig kubernetes.Interface, namespace string, cfg *config.MySQLConfig, persistPool *config.ConnectionPool) (sqlbuilder.Database, error) {
	if cfg.TableName == "" {
		return nil, errors.InternalError("tableName is empty")
	}

	ctx := context.Background()
	userNameByte, err := util.GetSecrets(ctx, kubectlConfig, namespace, cfg.UsernameSecret.Name, cfg.UsernameSecret.Key)
	if err != nil {
		return nil, err
	}
	passwordByte, err := util.GetSecrets(ctx, kubectlConfig, namespace, cfg.PasswordSecret.Name, cfg.PasswordSecret.Key)
	if err != nil {
		return nil, err
	}

	session, err := mysql.Open(mysql.ConnectionURL{
		User:     string(userNameByte),
		Password: string(passwordByte),
		Host:     cfg.GetHostname(),
		Database: cfg.Database,
		Options:  cfg.Options,
	})
	if err != nil {
		return nil, err
	}
	session = ConfigureDBSession(session, persistPool)
	// this is needed to make MySQL run in a Golang-compatible UTF-8 character set.
	_, err = session.Exec("SET NAMES 'utf8mb4'")
	if err != nil {
		return nil, err
	}
	_, err = session.Exec("SET CHARACTER SET utf8mb4")
	if err != nil {
		return nil, err
	}
	return session, nil
}

// ConfigureDBSession configures the DB session
func ConfigureDBSession(session sqlbuilder.Database, persistPool *config.ConnectionPool) sqlbuilder.Database {
	if persistPool != nil {
		session.SetMaxOpenConns(persistPool.MaxOpenConns)
		session.SetMaxIdleConns(persistPool.MaxIdleConns)
		session.SetConnMaxLifetime(time.Duration(persistPool.ConnMaxLifetime))
	}
	return session
}
