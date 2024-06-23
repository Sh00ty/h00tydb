package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"gitlab.com/Sh00ty/hootydb/internal/dbinit"
)

var (
	env = os.Getenv("ENV")
)

type db struct {
	name  string
	color *color.Color
}

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Kill, os.Interrupt)
	defer cancel()

	err := godotenv.Load()
	if err != nil {
		log.Panic("failed to parse dotenv")
	}
	dbs := []db{
		{name: "DB1", color: color.New(color.FgCyan)},
		{name: "DB2", color: color.New(color.FgHiYellow)},
		{name: "DB3", color: color.New(color.FgHiMagenta)},
	}
	for _, db := range dbs {
		dbinit.StartDb(ctx, env, db.name, db.color.Sprint(db.name))
	}
	<-ctx.Done()
}
