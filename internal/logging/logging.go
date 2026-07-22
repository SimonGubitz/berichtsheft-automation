package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"
)

type LogEnvMode = string

const (
	FileOnly LogEnvMode = "file"
	Test     LogEnvMode = "test"
	Dev      LogEnvMode = "development"
	Prod     LogEnvMode = "release"
)

func SetupDefault(envMode LogEnvMode) (io.Closer, error) {

	var logHandler slog.Handler
	closer := io.NopCloser(nil)

	switch envMode {
	case FileOnly:
		logFile, err := getLogFile()
		if err != nil {
			return nil, err
		}

		logWriter := logFile
		logHandler = slog.NewJSONHandler(logWriter, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})

		closer = logFile
	case Dev:
		// Set up a MultiWriter to stdout and the log file
		logFile, err := getLogFile()
		if err != nil {
			return nil, err
		}

		logWriter := io.MultiWriter(os.Stdout, logFile)
		logHandler = slog.NewJSONHandler(logWriter, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})

		closer = logFile

	case Test:
		logWriter := os.Stdout
		logHandler = slog.NewJSONHandler(logWriter, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})

	case Prod:
		logWriter := os.Stdout
		logHandler = slog.NewJSONHandler(logWriter, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	slog.SetDefault(slog.New(logHandler))
	return closer, nil
}

func getLogFile() (*os.File, error) {
	if err := os.MkdirAll(".logs", 0755); err != nil {
		return nil, fmt.Errorf("failed to create `.logs` directory, err: %w", err)
	}
	logFile, err := os.OpenFile(fmt.Sprintf(".logs/%s.log", time.Now().Format("20060102_150405")), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file, err: %w", err)
	}

	return logFile, nil
}
