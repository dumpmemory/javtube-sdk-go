package main

import (
	goflag "flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/peterbourgon/ff/v3"

	"github.com/javtube/javtube-sdk-go/database"
	"github.com/javtube/javtube-sdk-go/engine"
	V "github.com/javtube/javtube-sdk-go/internal/constant"
	"github.com/javtube/javtube-sdk-go/route"
	"github.com/javtube/javtube-sdk-go/route/validator"
)

const defaultRequestTimeout = 2 * time.Minute

var (
	opts = new(options)
	flag = goflag.NewFlagSet("", goflag.ExitOnError)
)

type options struct {
	// base options
	bind  string
	port  string
	token string
	dsn   string

	// database options
	dbMaxIdleConns int
	dbMaxOpenConns int
	dbAutoMigrate  bool
	dbPreparedStmt bool

	// version flag
	versionFlag bool
}

func init() {
	// gin initiate
	gin.DisableConsoleColor()

	// flag parsing
	flag.StringVar(&opts.bind, "bind", "", "Bind address of server")
	flag.StringVar(&opts.port, "port", "8080", "Port number of server")
	flag.StringVar(&opts.token, "token", "", "Token to access server")
	flag.StringVar(&opts.dsn, "dsn", "", "Database Service Name")
	flag.IntVar(&opts.dbMaxIdleConns, "db-max-idle-conns", 0, "Database max idle connections")
	flag.IntVar(&opts.dbMaxOpenConns, "db-max-open-conns", 0, "Database max open connections")
	flag.BoolVar(&opts.dbAutoMigrate, "db-auto-migrate", false, "Database auto migration")
	flag.BoolVar(&opts.dbPreparedStmt, "db-prepared-stmt", false, "Database prepared statement")
	flag.BoolVar(&opts.versionFlag, "v", false, "Show version")
	ff.Parse(flag, os.Args[1:], ff.WithEnvVarNoPrefix())
}

func showVersionAndExit() {
	fmt.Printf("%s-%s\n",
		V.Version, V.GitCommit)
	os.Exit(0)
}

func main() {
	if opts.versionFlag {
		showVersionAndExit()
	}

	db, err := database.Open(&database.Config{
		DSN:                  opts.dsn,
		PreparedStmt:         opts.dbPreparedStmt,
		MaxIdleConns:         opts.dbMaxIdleConns,
		MaxOpenConns:         opts.dbMaxOpenConns,
		DisableAutomaticPing: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	app := engine.New(db, defaultRequestTimeout)
	if err = app.AutoMigrate(opts.dbAutoMigrate); err != nil {
		log.Fatal(err)
	}

	var token validator.Validator
	if opts.token != "" {
		token = validator.Token(opts.token)
	}

	var (
		addr   = net.JoinHostPort(opts.bind, opts.port)
		router = route.New(app, token)
	)
	if err = http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}
