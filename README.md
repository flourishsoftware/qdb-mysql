# qdb-mysql

Command line database query utility for MySQL.

## Dependencies

- Golang
- `mysqlsh` command line utility.

## Docs

Docs may be generated with the `docs` command.
- Outputs to the `docs` directory.

## Environment variables:

- MYSQL_SERVER
- MYSQL_USER
- MYSQL_PASS
- MYSQL_PORT

## Commands (Generate `docs` for more info):

1. `query`
- Queries `MYSQL_SERVER` using the query file (and line number) specified in the arguments.
- If the option to generate an output file is enabled, the output file will be by the same filename as the input file.
2. `session`
- Opens a database session with `MYSQL_SERVER`.
- All queries ran during the session are piped to a session file which is generated in the `.qdb-mysql` under your user's home directory.
- Session filenames are generated with a unique timestammp.
- While the session is open, you may run a query or queries from any file and/or line number.

## VSCode

This utility uses VSCode for some "glossy" functions.
- If you choose not to use VSCode, it is possible your editor can use the following features.

Features using VSCode:
1. Piping commands into the terminal based on keyboard shortcuts and selected line number.
- See `.vscode/keybindings.json`.
- These settings will need to be installed in you user settings. 
- Example installation location (MacOS): `/Users/<username>/Library/Application Support/Code/User/keybindings.json`.
1. Opening SQL results into a new editor tab.
- See `func openFile(file string) error` in `session.go`.
- This is done automatically after each query.
- If a tab is already opened

## Sub-Commands for `session` Command:

1. `q ${file}:${lineNumber}`
- Execute a query from a specific file `${file}`, starting at line number `${lineNumber}`.
1. `q+ ${file}:${lineNumber}`
- Same as above, but reopen the session file upon execution.
1. `q ${file}`
- Execute all queries query from a specific file `${file}`.
1. `q+ ${file}`
- Same as above, but reopen the session file upon execution.

## Noted Quirks

1. Leading comments are ran as empty queries instead of being extracted prior to execution.
1. When querying by line number, trailing comments can cause execution to continue. 
- To stop this, just use `;` without any comments when delineating queries.

## Contributions

You might find that this code needs some tweaks to run with your specific system or to meet your specific needs.
- PRs are welcome, and encouraged!