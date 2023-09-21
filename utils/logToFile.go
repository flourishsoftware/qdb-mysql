package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type LogrusFileHook struct {
	file      *os.File
	flag      int
	chmod     os.FileMode
	formatter *logrus.TextFormatter
}

func NewLogrusFileHook(file string, flag int, chmod os.FileMode) (*LogrusFileHook, error) {
	plainFormatter := &logrus.TextFormatter{DisableColors: true}
	logFile, err := os.OpenFile(file, flag, chmod)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to write file on filehook %v", err)
		return nil, err
	}

	return &LogrusFileHook{logFile, flag, chmod, plainFormatter}, err
}

// Fire event
func (hook *LogrusFileHook) Fire(entry *logrus.Entry) error {

	plainformat, err := hook.formatter.Format(entry)
	line := string(plainformat)
	_, err = hook.file.WriteString(line)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to write file on filehook(entry.String)%v", err)
		return err
	}

	return nil
}

func (hook *LogrusFileHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}

// SetupLogToFile sets up logging to std out and file simulataneously
func SetupLogToFile(baseFilenameNoExt string) error {

	// log output to file?
	// log the error here if there is one (for formatting)
	if len(baseFilenameNoExt) > 0 {

		// use this file extension
		// add a time-based suffix to the filename
		extentsion := "log"
		timeSuffix := time.Now().Format("2006_01_02_T_15_04_05")
		filename := fmt.Sprintf("%s_%s.%s", baseFilenameNoExt, timeSuffix, extentsion)
		logrus.Infof("LOG Filename: %s", filename)

		// remove the file
		// if it exists
		if FileExists(filename) {
			if err := os.Remove(filename); err != nil {
				// log the error here if there is one (for formatting)
				msg := "remove the old log file"
				logrus.WithError(err).Error(msg)
				return fmt.Errorf("%s: %w", msg, err)
			}
		}

		// Add the logging hook
		fileHook, err := NewLogrusFileHook(filename, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			msg := "could not remove the old log file"
			logrus.WithError(err).Error(msg)
			return fmt.Errorf("%s: %w", msg, err)
		}
		logrus.AddHook(fileHook)
	}

	return nil
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func DirExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}
