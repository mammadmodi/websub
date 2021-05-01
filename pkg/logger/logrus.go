package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// NewLogrusLogger is an factory function for logrus logger
func NewLogrusLogger(config Configuration) (*logrus.Logger, error) {
	l := logrus.New()
	if !config.Enabled {
		l.SetOutput(ioutil.Discard)
		return l, nil
	}

	formatter := &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "time",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "msg",
			logrus.FieldKeyFunc:  "caller",
			logrus.FieldKeyFile:  "go_file",
		},
		PrettyPrint: config.Pretty,
	}
	l.SetFormatter(formatter)
	l.SetReportCaller(true)

	// setting log level
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		return nil, fmt.Errorf("error while parsing log level `%v`, err: %v", config.Level, err)
	}
	l.SetLevel(level)

	if config.CoreFields != nil {
		l.AddHook(&defaultFieldHook{
			defaultFields: config.CoreFields,
		})
	}

	// setting log outputs
	outputs := []io.Writer{os.Stdout}

	if config.FileRedirectEnabled {
		fileName := fmt.Sprintf("%s/%s.log", config.FileRedirectPath, config.FileRedirectPrefix)
		basePath := filepath.Dir(fileName)
		if _, err := os.Stat(basePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("base path `%s` is not exist, error: %v", basePath, err)
		}

		file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			return nil, fmt.Errorf("error while openning file `%s` for logging, err: %v", fileName, err)
		}
		outputs = append(outputs, file)
		l.ExitFunc = func(i int) {
			if err = file.Close(); err != nil {
				println(fmt.Sprintf("error while closing log file: %v", err))
			}
			println(fmt.Sprintf("log file with address %s closed successfully", fileName))
		}
	}

	mw := io.MultiWriter(outputs...)
	l.SetOutput(mw)

	return l, nil
}

// defaultFieldHook adds the ability to set default fields to logrus logger
type defaultFieldHook struct {
	defaultFields map[string]interface{}
}

func (h *defaultFieldHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *defaultFieldHook) Fire(e *logrus.Entry) error {
	for k, v := range h.defaultFields {
		e.Data[k] = v
	}
	return nil
}
