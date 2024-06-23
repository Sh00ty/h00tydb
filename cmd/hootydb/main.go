package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/joho/godotenv"
	"gitlab.com/Sh00ty/hootydb/internal/dbinit"
)

var (
	env = os.Getenv("ENV")
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: hootydb db1")
		return
	}

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Kill, os.Interrupt)
	defer cancel()

	err := godotenv.Load()
	if err != nil {
		log.Panic("failed to parse dotenv")
	}

	dbName := os.Args[1]
	dbinit.StartDb(ctx, env, dbName, dbName)
	<-ctx.Done()
}
