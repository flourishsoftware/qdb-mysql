package main

import (
	"context"
	"fmt"
	"os"

	"qdb-mysql/cmd"

	"github.com/sirupsen/logrus"
	"github.com/subosito/gotenv"
)

func main() {

	// Load the ENV.
	if err := gotenv.OverLoad(".env"); err != nil {
		fmt.Println(err)
	}

	// Create the CLI.
	opts := &cmd.CLI{
		DBHost: os.Getenv("MYSQL_SERVER"),
		DBUser: os.Getenv("MYSQL_USER"),
		DBPass: os.Getenv("MYSQL_PASS"),
		DBPort: os.Getenv("MYSQL_PORT"),
	}

	// Specify verbosity.
	logrus.SetLevel(logrus.TraceLevel)

	// Execute the CLI.
	if err := opts.Execute(context.Background()); err != nil {
		logrus.Errorf(err.Error())
		os.Exit(1)
	}
}
