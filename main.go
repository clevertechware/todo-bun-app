package main

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/clevertechware/todo-bun-app/internal/cmd"
)

func main() {
	app := &cli.Command{
		Name:    "todo-app",
		Usage:   "A TODO application using Bun ORM and PostgreSQL",
		Version: "1.0.0",
		Commands: []*cli.Command{
			cmd.ServeCommand(),
			cmd.MigrateCommand(),
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
