package session

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"qdb-mysql/utils"
	"strconv"
	"strings"
)

type QueryFileCmd struct {
	stdin                      io.WriteCloser
	cmdText                    string
	resultFile                 *os.File
	resultFileName             string
	queryText                  string
	openResultFileAfterExecute bool
	FilePath                   string
	LineNumber                 int
}

var _ Cmd = &QueryFileCmd{}

func NewQueryFileCmd(stdin io.WriteCloser, cmdText string, openResultFileAfterExecute bool) *QueryFileCmd {
	return &QueryFileCmd{
		stdin:                      stdin,
		cmdText:                    cmdText,
		openResultFileAfterExecute: openResultFileAfterExecute,
	}
}

func (queryFromFile *QueryFileCmd) String() string {
	return fmt.Sprintf("q %s %d", queryFromFile.FilePath, queryFromFile.LineNumber)
}

func (queryFromFile *QueryFileCmd) QueryText() string {
	return queryFromFile.queryText
}

func (queryFromFile *QueryFileCmd) ResultFile() (bool, *os.File) {
	return queryFromFile.resultFile != nil, queryFromFile.resultFile
}

func (queryFromFile *QueryFileCmd) ResultFileName() string {
	return queryFromFile.resultFileName
}

func (queryFromFile *QueryFileCmd) OpenResultFileAfterExecute() bool {
	return queryFromFile.openResultFileAfterExecute
}

func (queryFromFile *QueryFileCmd) ParseArgs(sessionFile string) (err error) {

	// Since we already read, it remove the first chahracter, then re-trim.
	cutIndex := 1
	if queryFromFile.openResultFileAfterExecute {
		cutIndex++
	}
	args := queryFromFile.cmdText[cutIndex:]
	// fmt.Println(queryFromFile.cmdText, cutIndex, args)
	args = strings.TrimSpace(args)
	splitsOnColon := strings.Split(args, ":")
	if len(splitsOnColon) == 0 {
		return fmt.Errorf("invalid args: %s", queryFromFile.cmdText)
	} else if len(splitsOnColon) == 1 {
		queryFromFile.FilePath = splitsOnColon[0]
	} else if len(splitsOnColon) >= 2 {
		queryFromFile.FilePath = splitsOnColon[0]
		lineNumber, err := strconv.Atoi(splitsOnColon[1])
		if err != nil {
			return fmt.Errorf("failed to parse line number: %s (%w)", splitsOnColon[1], err)
		}
		queryFromFile.LineNumber = lineNumber
	}

	// FilePath is required for the input sql.
	if len(queryFromFile.FilePath) <= 0 {
		return fmt.Errorf("cannot tolerate empty input file name: %s", queryFromFile.FilePath)
	}

	// Validate the in-file extension.
	inFileExtension := filepath.Ext(queryFromFile.FilePath)
	if !strings.EqualFold(inFileExtension, ".sql") {
		return fmt.Errorf("[WARNING] file must have a `.sql` extension")
	}

	// Uncomment this to put a log file beside the sql file.
	// // Open log file for the in-file.
	// inFileDir, inFilePath := filepath.Split(queryFromFile.FilePath)
	// resultFilePath := strings.TrimSuffix(inFilePath, inFileExtension)
	// resultFilePath = resultFilePath + ".qbd"
	// resultFilePath = filepath.Join(inFileDir, resultFilePath)
	// var resultFile *os.File
	// if !utils.FileExists(resultFilePath) {
	// 	resultFile, err = os.Create(resultFilePath)
	// 	if err != nil {
	// 		return err
	// 	}
	// } else {
	// 	resultFile, err = os.OpenFile(resultFilePath, os.O_WRONLY|os.O_APPEND, 0644)
	// 	if err != nil {
	// 		return fmt.Errorf("error opening result file: %s (%w)", resultFilePath, err)
	// 	}
	// }
	// if resultFile == nil {
	// 	return fmt.Errorf("cannot tolerate nil resultFile for path: %s", resultFilePath)
	// }
	// queryFromFile.resultFile = resultFile
	// queryFromFile.resultFileName = resultFile.Name()
	// DO NOT CLOSE THE FILE HERE :)
	// defer resultFile.Close()

	// Read the in-file.
	inBytes, err := os.ReadFile(queryFromFile.FilePath)
	if err != nil {
		return err
	}
	allSQLInFile := string(inBytes)
	queryFromFile.queryText = string(inBytes)

	// Print the action before executing.
	fmt.Printf(" %s Session in %s\n", responseBullet, sessionFile)
	if queryFromFile.LineNumber <= 0 {
		fmt.Printf(" %s Run file %s\n", responseBullet, queryFromFile.FilePath)
	} else {
		fmt.Printf(" %s Run file %s, line %d\n", responseBullet, queryFromFile.FilePath, queryFromFile.LineNumber)
	}
	// fmt.Printf(" %s Results in %s\n", responseBullet, queryFromFile.resultFileName)

	// Slice-n-dice the SQL.
	if queryFromFile.LineNumber <= 0 {
		numLines := len(strings.Split(queryFromFile.queryText, "\n"))
		fmt.Printf(".. %d lines\n", numLines)
		queryFromFile.queryText = utils.RemoveCommentLinesBetweenStatements(allSQLInFile)
		queryFromFile.queryText = utils.EnsureSQLEndsInSemiColon(queryFromFile.queryText)
		fmt.Println(queryFromFile.queryText)
	} else {
		lineNumberSQL, err := utils.ParseSQLAtLineNumber(allSQLInFile, queryFromFile.LineNumber)
		if err != nil {
			return err
		}
		queryFromFile.queryText = utils.RemoveCommentLinesBetweenStatements(lineNumberSQL)
		queryFromFile.queryText = utils.EnsureSQLEndsInSemiColon(queryFromFile.queryText)
		fmt.Println(queryFromFile.queryText)
	}

	return nil
}

func (queryFromFile *QueryFileCmd) Execute() error {
	// fmt.Println("queryFromFileCmd")

	// Validate we have a non-nil stdin.
	if queryFromFile.stdin == nil {
		return fmt.Errorf("cannot tolerate nil stdin")
	}

	// queryFromFile.queryText is required.
	if len(queryFromFile.queryText) == 0 {
		return fmt.Errorf("empty query")
	}

	// dispatch the sql to the mysqlsh hbinary
	queryFromFile.stdin.Write([]byte(queryFromFile.queryText))

	return nil
}
