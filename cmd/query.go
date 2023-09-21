package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"qdb-mysql/utils"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type QueryOpts struct {
	InFile                       string
	ExecuteStatementAtLineNumber int
	CreateNewFile                bool
	DoNotReopenResultFile        bool
	// UseMysqlSh                   bool
}

var (
	queryCmdOpts = QueryOpts{}
	queryCmd     = &cobra.Command{
		Use:     "query",
		Short:   "Query a MySQL DB.",
		Long:    ``,
		Example: ``,
		Args:    cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			if err := cli.Setup(); err != nil {
				return err
			}
			defer cli.Teardown()

			return QueryMysqlCmdE(cli, queryCmdOpts)
		},
	}
)

func init() {
	rootCmd.AddCommand(queryCmd)
	queryCmd.Flags().StringVarP(&queryCmdOpts.InFile, "file", "f", "query-something.sql", "Filepath of the sql script to execute.")
	queryCmd.Flags().BoolVarP(&queryCmdOpts.CreateNewFile, "new", "n", false, "Set this flag to write a new output file on every run.")
	queryCmd.Flags().IntVarP(&queryCmdOpts.ExecuteStatementAtLineNumber, "line", "l", -1, "Specify the line number of the statement to run.")
	queryCmd.Flags().BoolVarP(&queryCmdOpts.DoNotReopenResultFile, "silent", "s", false, "Specify this flag to not open the result file after running the command.")
	// queryCmd.Flags().BoolVarP(&queryCmdOpts.UseMysqlSh, "mysqlsh", "m", false, "Specify this flag to use mysqlsh instead of mysql.")
}

// QueryMysqlCmdE runs a command
func QueryMysqlCmdE(c *CLI, opts QueryOpts) (err error) {

	fmt.Println("In File:", opts.InFile)
	if strings.TrimRight(strings.ToLower(opts.InFile), ".sql") == strings.ToLower(opts.InFile) {
		return fmt.Errorf("in-file must have a `.sql` extension")
	}

	// Read the SQL from the in-File.
	inBytes, err := os.ReadFile(opts.InFile)
	if err != nil {
		return err
	}
	inSQL := string(inBytes)

	// Should we run the whole file, or just one line?
	if opts.ExecuteStatementAtLineNumber > 0 {
		inSQL, err = utils.ParseSQLAtLineNumber(inSQL, opts.ExecuteStatementAtLineNumber)
		if err != nil {
			return err
		}
		fmt.Println(inSQL)
	}

	var terminalCmd utils.BashCmd
	// if opts.UseMysqlSh {
	terminalCmd = utils.NewBashCmd("mysqlsh").
		Append(fmt.Sprintf("-u%s", c.DBUser)).
		Append(fmt.Sprintf("-p%s", c.DBPass)).
		Append(fmt.Sprintf("-h%s", c.DBHost)).
		Append(fmt.Sprintf("-P%s", c.DBPort)).
		Append("--ssl-mode=DISABLED"). // Required
		Append("--sql").
		// Append("--verbose=1").
		Append("--log-level=info").
		// Append("--dba-log-sql=2").
		// Append("--show-warnings=true").
		// Append("--result-format=json/array").
		Append("--result-format=table").
		// Append("--result-format=json/pretty").
		// Append("--file").Append(opts.InFile)
		Append("--interactive").
		Append("--execute").Append(inSQL)
	// Append("--execute").Append("SET NAMES utf8; SET autocommit = 0;" + "\n" + inSQL)
	// } else {
	// 	terminalCmd = utils.NewBashCmd("mysql").
	// 		Append(fmt.Sprintf("-u%s", c.DBUser)).
	// 		Append(fmt.Sprintf("-p%s", c.DBPass)).
	// 		Append(fmt.Sprintf("-h%s", c.DBHost)).
	// 		Append(fmt.Sprintf("-P%s", c.DBPort)).
	// 		Append("--show-warnings=true").
	// 		// Append("--comments=true").
	// 		// Append("--reconnect"). // This helps with transactions?
	// 		// Append("--init-command").Append("SET NAMES utf8; SET autocommit = 0;").
	// 		// Append("--silent").
	// 		Append("-t").
	// 		// Append("--xml").
	// 		// Append("--html").
	// 		Append("-vv").
	// 		// Append("-e").Append(fmt.Sprintf("source %s", opts.InFile))
	// 		Append("-e").Append(`\` + "\n" + inSQL)
	// }

	cmdMain, cmdOpts := terminalCmd.AsCmd()
	cmd := exec.Command(cmdMain, cmdOpts...)
	// fmt.Println(cmd.String())
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	// cmd.Stdin = inFile
	output := ""
	// out, err := cmd.CombinedOutput()
	err = cmd.Run()
	if err != nil {
		// logrus.WithError(err).WithField("Stdout", fmt.Errorf(out.String())).WithField("Stderr", stderr.String()).Errorf("MySQL error")
		errOut := stderr.String()
		errOut = strings.ReplaceAll(errOut, "[33mWARNING: [0mUsing a password on the command line interface can be insecure.", "") // mysqlsh
		errOut = strings.ReplaceAll(errOut, "mysql: [Warning] Using a password on the command line interface can be insecure.", "")  // mysql
		if len(stdout.String()) > 0 {
			output = stdout.String() + "\n" + errOut
		} else {
			return fmt.Errorf("%w: %s", err, errOut)
		}
	} else {
		// output := string(out)
		output = stdout.String()
		// output = strings.ReplaceAll(output, "mysql: [Warning] Using a password on the command line interface can be insecure.", "")
		// output = strings.ReplaceAll(output, "\t\t", "\t \t")
		// output = strings.ReplaceAll(output, "\n", "?\n?")
	}
	// logrus.Info(output)

	/////////////////////////////////////////////////////////////////////////////////////

	//  Log everything to file.
	queryFileDir, filename := filepath.Split(opts.InFile)
	extension := filepath.Ext(filename)
	baseFilename := strings.TrimSuffix(filename, extension)

	if opts.CreateNewFile {
		timeSuffix := time.Now().Format("2006_01_02_T_15_04_05")
		baseFilename += fmt.Sprintf("_%s", timeSuffix)
	}

	// use this file extension
	// add a time-based suffix to the filename
	// outFilename := fmt.Sprintf("%s_%s.%s", baseFilename, timeSuffix, extentsion)
	outFilename := fmt.Sprintf("%s%s", baseFilename, ".qbd") // extension = ".sql"
	// outFilename = filepath.Join(outDir, outFilename)
	outFilename = filepath.Join(queryFileDir, outFilename)

	/////////////////////////////////////////////////////////////////////////////////////

	// Use os.Create to create a file for writing.
	// Create a new writer.
	// Write a string to the file.
	// Flush.
	f6, err := os.Create(outFilename)
	if err != nil {
		return err
	}
	w6 := bufio.NewWriter(f6)
	w6.WriteString(output)
	w6.Flush()
	fmt.Println("Out File:", outFilename)

	// Bail if we don't need to open the file.
	if opts.DoNotReopenResultFile {
		return nil
	}

	openFileErr := utils.OpenFileInEditor(outFilename)
	if openFileErr != nil {
		return openFileErr
	}

	return nil
}
