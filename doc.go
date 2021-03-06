/*
Package gue implements Golang queue on top of PostgreSQL.
It uses transaction-level locks for concurrent work.

PostgreSQL drivers

Package supports several PostgreSQL drivers using adapter interface internally.
Currently, adapters for the following drivers have been implemented:
 - github.com/jackc/pgx v4
 - github.com/jackc/pgx v3
 - github.com/lib/pq

Usage

Here is a complete example showing worker setup for pgx/v4 and two jobs enqueued, one with a delay:

	package main

	import (
		"context"
		"encoding/json"
		"fmt"
		"log"
		"os"
		"time"

		"github.com/jackc/pgx/v4/pgxpool"

		"github.com/vgarvardt/gue/v2"
		"github.com/vgarvardt/gue/v2/adapter/pgxv4"
	)

	type printNameArgs struct {
		Name string
	}

	func main() {
		printName := func(j *gue.Job) error {
			var args printNameArgs
			if err := json.Unmarshal(j.Args, &args); err != nil {
				return err
			}
			fmt.Printf("Hello %s!\n", args.Name)
			return nil
		}

		pgxCfg, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
		if err != nil {
			log.Fatal(err)
		}

		pgxPool, err := pgxpool.ConnectConfig(context.Background(), pgxCfg)
		if err != nil {
			log.Fatal(err)
		}
		defer pgxPool.Close()

		poolAdapter := pgxv4.NewConnPool(pgxPool)

		gc := gue.NewClient(poolAdapter)
		wm := gue.WorkMap{
			"PrintName": printName,
		}
		// create a pool w/ 2 workers
		workers := gue.NewWorkerPool(gc, wm, 2, gue.WithPoolQueue("name_printer"))

		ctx, shutdown := context.WithCancel(context.Background())

		// work jobs in goroutine
		if err := workers.Start(ctx); err != nil {
			log.Fatal(err)
		}

		args, err := json.Marshal(printNameArgs{Name: "vgarvardt"})
		if err != nil {
			log.Fatal(err)
		}

		j := &gue.Job{
			Type:  "PrintName",
			Args:  args,
		}
		if err := gc.Enqueue(context.Background(), j); err != nil {
			log.Fatal(err)
		}

		j := &gue.Job{
			Type:  "PrintName",
			RunAt: time.Now().UTC().Add(30 * time.Second), // delay 30 seconds
			Args:  args,
		}
		if err := gc.Enqueue(context.Background(), j); err != nil {
			log.Fatal(err)
		}

		time.Sleep(30 * time.Second) // wait for while

		// send shutdown signal to worker
		shutdown()
	}

*/
package gue
