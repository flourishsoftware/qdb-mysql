package cmd

// This file uses mysqlsh directly, and directs the stdout and stderr to file
//     descriptors while leaviing the stdin open the internal workings of the cli.
// NOTE: DID NOT WORK WELL WITH mysql, only caught the first statement-ish... about 40096 characters. Maybe a bash thing.

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"qdb-mysql/session"
	"qdb-mysql/utils"

	"path/filepath"
	"strings"
	"time"

	"github.com/hpcloud/tail"
	"github.com/spf13/cobra"
)

type SessionOpts struct{}

var (
	sessionCmdOpts = SessionOpts{}
	sessionCmd     = &cobra.Command{
		Use:     "session",
		Short:   "Query a MySQL DB inside of a session.",
		Long:    ``,
		Example: ``,
		Args:    cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			if err := cli.Setup(); err != nil {
				return err
			}
			defer cli.Teardown()

			return SessionMysqlCmdE(cli, sessionCmdOpts)
		},
	}
)

func init() {
	rootCmd.AddCommand(sessionCmd)
}

// SessionMysqlCmdE runs a command
func SessionMysqlCmdE(c *CLI, opts SessionOpts) (err error) {

	// Fetch the absolute session directory path.
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Setup the session file.
	timeFormatted := time.Now().Format("2006_01_02_T_15_04_05")
	sessionID := timeFormatted
	sessionFilePath := fmt.Sprintf(".session.%s.qdb", sessionID)
	sessionDir := filepath.Join(userHomeDir, ".qdb-mysql")
	sessionFilePath = filepath.Join(sessionDir, sessionFilePath)
	fmt.Println("Session File", sessionFilePath)
	if !utils.DirExists(sessionDir) {
		err = os.MkdirAll(sessionDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// Ensure the custom directory, if provided.
	if len(sessionFilePath) > 0 {
		err = os.MkdirAll(filepath.Dir(sessionFilePath), os.ModePerm)
		if err != nil {
			return err
		}
	}

	// Create new file, or Truncate existing file.
	sessionFile, err := os.Create(sessionFilePath)
	if err != nil {
		return err
	}
	defer sessionFile.Close()

	terminalCmd := utils.NewBashCmd("mysqlsh").
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
		Append("--quiet-start")
	// Append("--execute").Append(inSQL)
	// Append("--execute").Append("SET NAMES utf8; SET autocommit = 0;" + "\n" + inSQL)

	cmdMain, cmdOpts := terminalCmd.AsCmd()
	cmd := exec.Command(cmdMain, cmdOpts...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	cmd.Stdout = sessionFile
	cmd.Stderr = sessionFile

	// Defer the closing of the result files.
	cmdsByResultFileName := make(map[string]session.Cmd)
	defer func() {
		for _, scannerCommand := range cmdsByResultFileName {
			ok, resultFile := scannerCommand.ResultFile()
			if !ok || resultFile == nil {
				continue
			}
			if err := resultFile.Close(); err != nil {
				fmt.Printf("failed to close result file: %s\n", resultFile.Name())
				return
			}
		}
	}()

	// Tail logs to session file... As tailed logs come in, also print them to the terminal.
	t, err := tail.TailFile(
		sessionFilePath, tail.Config{Follow: true, ReOpen: true})
	if err != nil {
		// done <- nil
		fmt.Println("Session Tail Error: %w", err)
	}
	defer func() {
		fmt.Println("kill tail")
		t.Kill(nil)
	}()
	go func() {
		for line := range t.Lines {
			fmt.Println(line.Text)
		}
	}()

	// Start the session.
	err = cmd.Start()
	if err != nil {
		return err
	}

	// Stream / tail the session output file.
	// Every time we run a statement in a new file, it will have
	//     the output directed to an output file nearby the input.

	// Listen to events from the user's keyboard.
	heartUtf8 := "\u2665"
	// responseBullet := "--->"
	listBullet := ">"
	displayKeyboardCommandSuggestions := func() {
		fmt.Printf("\n")
		fmt.Printf("Type `exit` to quit.\n")
		fmt.Printf("Enter `qdb-mysql` command:\n")
		fmt.Printf("  %s q test.sql    = Run the whole file.\n", listBullet)
		fmt.Printf("  %s q test.sql:10 = Run line 10 from the file.\n", listBullet)
		// fmt.Printf("  %s select 1;	   = Run a raw query.\n", listBullet) // Used to pipe in a selection of text.
	}
	displayKeyboardCommandSuggestions()

	// Scan for each successive command.
	scanner := bufio.NewScanner(os.Stdin)
	for {

		// Wait until the output is initialized.
		fmt.Printf("\n")
		fmt.Printf("\n %s> ", heartUtf8)

		if scanner.Scan() {
			// stdin.Write([]byte("select 1;\n"))

			// Handle the text inputted.
			text := scanner.Text()
			text = strings.TrimSpace(text)
			text = strings.ReplaceAll(text, "'", "")

			// Break when empty scanned text.
			if len(text) == 0 {
				fmt.Println("Try again!")
				displayKeyboardCommandSuggestions()
				continue
			}

			if strings.EqualFold(text, "exit") || strings.EqualFold(text, "x") {
				fmt.Println("Bye!")
				// done <- nil
				break
			}

			// Get the session command from the scanner text.
			scannerCmd, err := session.GetSessionCmd(text, stdin)
			if err != nil {
				fmt.Printf("Command Error: %s\n", err.Error())
				displayKeyboardCommandSuggestions()
				continue
			}

			// Parse the command args from the scanner text.
			if err := scannerCmd.ParseArgs(sessionFilePath); err != nil {
				fmt.Printf("Parse Error: %s\n", err.Error())
				displayKeyboardCommandSuggestions()
				continue
			}

			// Execute the scanner command.
			if err := scannerCmd.Execute(); err != nil {
				fmt.Printf("Exec Error: %s\n", err.Error())
			}

			// Open the file again after execution. Let the editor figure out if it is already opened.
			if scannerCmd.OpenResultFileAfterExecute() {
				_ = utils.OpenFileInEditor(sessionFilePath)
				// _ = openFile(userHomeDir, opts.VsCodeWorkspaceFile, scannerCmd.ResultFileName())
			}
		} else {
			break
		}
	}
	return cmd.Wait()
}
