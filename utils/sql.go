package utils

import (
	"fmt"
	"strings"
)

func trimSqlLine(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "\n\t")
	s = strings.TrimSuffix(s, "\n")
	return s
}

func RemoveCommentLinesBetweenStatements(inSQL string) string {
	sqlLineDelimiter := "\n"
	sqlLines := strings.Split(inSQL, sqlLineDelimiter)
	for i := 0; i < len(sqlLines); i++ {
		sqlLine := sqlLines[i]
		sqlLineTrimmed := trimSqlLine(sqlLine)
		isComment := strings.Index(sqlLineTrimmed, "-- ") == 0
		if isComment {
			sqlLines = append(sqlLines[:i], sqlLines[i+1:]...)
		}
	}
	return strings.Join(sqlLines, "\n")
}

func ParseSQLAtLineNumber(inSQL string, lineNumber int) (string, error) {

	// Build a map of all the sql lines.
	sqlLineDelimiter := "\n"
	sqlLines := strings.Split(inSQL, sqlLineDelimiter)
	sqlLinesByLineNum := make(map[int]string)
	for i, sqlLine := range sqlLines {
		sqlLinesByLineNum[i+1] = sqlLine
	}

	// Find the start of the statement.
	// Start with the requested lineNumber.
	foundBeginningOfStatement := false
	beginningCursor := lineNumber
	preSQL := ""
	for {
		beginningCursor--

		// Examine the line before the last.
		cursorSqlLine, ok := sqlLinesByLineNum[beginningCursor]
		if !ok {
			break
		}
		cursorSqlLineTrimmed := trimSqlLine(cursorSqlLine)
		if strings.HasPrefix(cursorSqlLineTrimmed, `\`) {
			cursorSqlLine = strings.Replace(cursorSqlLine, `\`, "", 1)
		}

		// If it contains any semicolons, take all the text after the last colon.
		lastIndexOfSemiColon := strings.LastIndex(cursorSqlLineTrimmed, ";")
		if lastIndexOfSemiColon >= 0 && lastIndexOfSemiColon < len(cursorSqlLineTrimmed)-1 {
			preSQL = cursorSqlLine[lastIndexOfSemiColon:] + sqlLineDelimiter + preSQL
			foundBeginningOfStatement = true
		} else if lastIndexOfSemiColon >= 0 && lastIndexOfSemiColon >= len(cursorSqlLineTrimmed)-1 {
			foundBeginningOfStatement = true
		} else {
			preSQL = cursorSqlLine + sqlLineDelimiter + preSQL
		}

		// If we found the beginning of the SQL statement, then break the loop.
		if foundBeginningOfStatement {
			break
		}
	}

	// Go to the line number in question.
	sqlLine, ok := sqlLinesByLineNum[lineNumber]
	if !ok {
		return "", fmt.Errorf("line number %d does not exist", lineNumber)
	}
	sqlLineTrimmed := trimSqlLine(sqlLine)

	// Do we need to check the following lines to see if the statement ends here?
	// Does this line end with a semicolon?
	// If so, is it a comment?
	keepCheckingForStatementEnd := true
	endsWithSemiColon := strings.LastIndex(sqlLineTrimmed, ";") >= len(sqlLineTrimmed)-1
	isComment := strings.Index(sqlLineTrimmed, "-- ") == 0
	isEmpty := len(sqlLineTrimmed) == 0
	// fmt.Println("sqlLine", sqlLine)
	// fmt.Println("endsWithSemiColon", endsWithSemiColon)
	// fmt.Println("isComment", isComment)
	// fmt.Println("isEmpty", isEmpty)
	if endsWithSemiColon && !isComment && !isEmpty {
		// fmt.Println("you")
		keepCheckingForStatementEnd = false
	}

	// Find the end of the statement.
	// Maybe the end of the line has a `;` in it!
	foundEndingOfStatement := false
	endingCursor := lineNumber
	postSQL := ""
	if keepCheckingForStatementEnd {
		for {
			endingCursor++

			// Examine the line after the last, beginning with the line in question.
			cursorSqlLine, ok := sqlLinesByLineNum[endingCursor]
			if !ok {
				break
			}
			cursorSqlLineTrimmed := trimSqlLine(cursorSqlLine)

			// Find the first semicolon and stop there.
			firstIndexOfSemiColon := strings.Index(cursorSqlLine, ";")
			isComment := strings.Index(cursorSqlLineTrimmed, "-- ") == 0
			if !isComment && firstIndexOfSemiColon > -1 {
				postSQL = postSQL + sqlLineDelimiter + cursorSqlLine[0:firstIndexOfSemiColon]
				foundEndingOfStatement = true
			} else {
				postSQL = postSQL + sqlLineDelimiter + cursorSqlLine
			}

			// If we found the ending of the SQL statement, then break the loop.
			if foundEndingOfStatement {
				break
			}
		}
	}

	// Build the out SQL, and trim it.
	outSQL := trimSqlLine(preSQL + sqlLine + postSQL)

	// fmt.Println(outSQL)
	return EnsureSQLEndsInSemiColon(outSQL), nil
}

func EnsureSQLEndsInSemiColon(sql string) string {
	// If it does not end with a semicolon and line break, do so!
	if !strings.HasSuffix(sql, "\n") {
		sql += "\n"
	}
	if !strings.HasSuffix(sql, ";\n") {
		sql = sql[0 : len(sql)-1]
		sql += ";\n"
	}
	return sql
}
