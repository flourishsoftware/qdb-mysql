package cmd

import (
	"context"

	"qdb-mysql/utils"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
)

var (
	version = "1.0.0"
)

// CLI are the global args
type CLI struct {
	DBHost, DBUser, DBPass, DBPort string
	Version                        string
	context                        context.Context
	cancelFn                       context.CancelFunc
}

// Teardown tears down connections
func (ropts *CLI) Teardown() {

	// cancel context
	if cli.cancelFn != nil {
		cli.cancelFn()
	}
}

// Setup initializes connections to MySQL, MongoDB, and METRC
func (ropts *CLI) Setup() (err error) {

	// logrus.Infof("%+v", ropts)
	// set up background context and defer cancel untl end of main
	cli.context, cli.cancelFn = context.WithCancel(cli.context)

	return nil
}

var (
	cli     = &CLI{}
	rootCmd = &cobra.Command{
		Use:   "qdb-mysql",
		Short: "Query the DB",
		Long:  `Making things easier!! `,
		// Run:   func(cmd *cobra.Command, args []string) {},
	}
)

func init() {
}

// Execute starts the CLI.
func (c *CLI) Execute(ctx context.Context) error {

	// Set the global cli & details
	c.context = ctx
	cli = c

	// start the cli
	return rootCmd.Execute()
}

// SetupLogToFile sets up logging to std out and file simulataneously
func (c *CLI) SetupLogToFile(baseFilenameNoExt string) error {
	return utils.SetupLogToFile(baseFilenameNoExt)
}
