package main

import (
	"database/sql"
	"log"
	"mailinglist/grpcapi"
	"mailinglist/jsonapi"
	"mailinglist/mdb"
	"sync"

	"github.com/alexflint/go-arg"
)

var args struct {
	// We can specify the mailinglist_db in the commandline, otherwise it is set to a default
	DbPath string `arg:"env:MAILINGLIST_DB`
	BindJson string `arg:"env:MAILINGLIST_BIND_JSON"`
	BindGrpc string `arg:"env:MAILINGLIST_BIND_GRPC"`
}

func main() {
	arg.MustParse(&args)

	if args.DbPath == "" {
		args.DbPath = "list.db" // default DB location
	}
	if args.BindJson == "" {
		args.BindJson = ":8080" // default port
	}
	if args.BindGrpc == "" {
		args.BindGrpc = ":8081" // default port
	}

	log.Printf("using database '%v'\n", args.DbPath)
	db, err := sql.Open("sqlite3", args.DbPath)

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mdb.TryCreate(db)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		log.Printf("starting JSON API server...\n")
		jsonapi.Serve(db, args.BindJson)
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		log.Printf("starting gRPC API server...\n")
		grpcapi.Serve(db, args.BindGrpc)
		wg.Done()
	}()

	wg.Wait()
}

