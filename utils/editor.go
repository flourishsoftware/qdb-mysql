package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"
)

func OpenFileInEditor( /* userHomeDir string, workspaceFile string, */ file string) error {
	<-time.After(time.Millisecond * 750)
	// dataDir := filepath.Join(userHomeDir, ".vscode-qbd")
	openCmd := NewBashCmd("code").
		// Append("--user-data-dir").Append(dataDir).
		Append("-r")
		// Append("-g102").
	// if len(workspaceFile) > 0 {
	// 	openCmd = openCmd.Append(workspaceFile)
	// }
	openCmd = openCmd.Append("--goto").Append(file + ":" + "999999999999")
	// openCmd = openCmd.Append(file)
	openCmdMain, openCmdOpts := openCmd.AsCmd()
	editorCmd := exec.Command(openCmdMain, openCmdOpts...)
	// logrus.Info(editorCmd.String())
	var editorCmdStdErr bytes.Buffer
	editorCmd.Stderr = &editorCmdStdErr
	if err := editorCmd.Run(); err != nil {
		// logrus.WithField("error", editorCmdStdErr.String()).Error("error writing final file")
		return fmt.Errorf(editorCmdStdErr.String())
	}
	return nil
}
