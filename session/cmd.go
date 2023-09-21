package session

import (
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	responseBullet = "√"
)

type Cmd interface {
	Execute() error
	String() string
	OpenResultFileAfterExecute() bool
	ResultFile() (bool, *os.File)
	ResultFileName() string
	QueryText() string
	ParseArgs(sessionFile string) error
}

func GetSessionCmd(text string, stdin io.WriteCloser) (cmd Cmd, err error) {
	trimmed := strings.TrimSpace(text)

	// Can't be empty
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("not enough args")
	}

	// Get the first part of the command.
	splitsOnSpace := strings.Split(trimmed, " ")
	if len(splitsOnSpace) == 0 || len(splitsOnSpace[0]) == 0 {
		return nil, fmt.Errorf("not enough args")
	}
	baseCommand := splitsOnSpace[0]

	// Match the first part to determine the command.
	switch strings.ToLower(baseCommand) {
	case "q":
		cmd = NewQueryFileCmd(stdin, trimmed, false)
	case "q+":
		cmd = NewQueryFileCmd(stdin, trimmed, true)
	// case "g":
	// 	// TODO
	// 	cmd = NewSessionGenerateCmd(stdin, trimmed, true)
	// 	return nil, fmt.Errorf("invalid command: %s", trimmed)
	default:
		cmd = NewQueryRawCmd(stdin, trimmed)
		// return nil, fmt.Errorf("invalid command: %s", trimmed)
	}

	return cmd, nil
}
