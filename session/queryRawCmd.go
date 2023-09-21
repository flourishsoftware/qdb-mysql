package session

import (
	"fmt"
	"io"
	"os"
	"qdb-mysql/utils"
)

type QueryRawCmd struct {
	stdin     io.WriteCloser
	cmdText   string
	queryText string
}

var _ Cmd = &QueryRawCmd{}

func NewQueryRawCmd(stdin io.WriteCloser, cmdText string) *QueryRawCmd {
	return &QueryRawCmd{
		stdin:   stdin,
		cmdText: cmdText,
	}
}

func (rawQuery *QueryRawCmd) String() string {
	return rawQuery.queryText
}

func (rawQuery *QueryRawCmd) QueryText() string {
	return rawQuery.queryText
}

func (rawQuery *QueryRawCmd) ResultFile() (bool, *os.File) {
	return false, nil
}

func (rawQuery *QueryRawCmd) ResultFileName() string {
	return "NO_FILE"
}

func (rawQuery *QueryRawCmd) OpenResultFileAfterExecute() bool {
	return false
}

func (rawQuery *QueryRawCmd) ParseArgs(sessionFile string) (err error) {
	queryText := utils.RemoveCommentLinesBetweenStatements(rawQuery.cmdText)
	rawQuery.queryText = utils.EnsureSQLEndsInSemiColon(queryText)
	rawQuery.cmdText = queryText

	// Print the action before executing.
	fmt.Printf(" %s Session in %s\n", responseBullet, sessionFile)
	fmt.Printf(" %s Run raw query: %s\n", responseBullet, queryText)

	return nil
}

func (rawQuery *QueryRawCmd) Execute() error {
	// fmt.Println("rawQueryCmd")

	// Validate we have a non-nil stdin.
	if rawQuery.stdin == nil {
		return fmt.Errorf("cannot tolerate nil stdin")
	}

	// rawQuery.queryText is required.
	if len(rawQuery.queryText) == 0 {
		return fmt.Errorf("empty query")
	}

	// dispatch the sql to the mysqlsh hbinary
	rawQuery.stdin.Write([]byte(rawQuery.queryText))

	return nil
}
