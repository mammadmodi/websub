package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestNewLogrusLogger(t *testing.T) {
	t.Run("test when logger is not enabled", func(t *testing.T) {
		c := Configuration{
			Enabled: false,
		}
		lgrus, err := NewLogrusLogger(c)
		assert.NoError(t, err)
		assert.Equal(t, ioutil.Discard, lgrus.Out)
	})

	t.Run("test when level is not valid", func(t *testing.T) {
		c := Configuration{
			Enabled: true,
			Level:   "invalid",
		}
		lgrus, err := NewLogrusLogger(c)
		assert.Error(t, err)
		assert.Nil(t, lgrus)
	})

	t.Run("testing when logger is expected to log into a non exist path", func(t *testing.T) {
		c := Configuration{
			Enabled:             true,
			Level:               "info",
			FileRedirectEnabled: true,
			FileRedirectPath:    "./not_exist",
		}
		l, err := NewLogrusLogger(c)
		assert.Error(t, err)
		assert.Nil(t, l)
	})

	t.Run("testing logger will continue when syslog server has error", func(t *testing.T) {
		c := Configuration{
			Enabled: true,
			Level:   "info",
		}
		l, err := NewLogrusLogger(c)
		assert.NoError(t, err)
		assert.NotNil(t, l)
	})

	t.Run("testing when logger is expected to log into file", func(t *testing.T) {
		c := Configuration{
			Enabled:             true,
			Level:               "info",
			FileRedirectEnabled: true,
			FileRedirectPath:    ".",
			FileRedirectPrefix:  "webis_test_log",
			CoreFields:          map[string]interface{}{"key": "default_value"},
		}

		lgrus, err := NewLogrusLogger(c)
		assert.NoError(t, err)
		assert.Equal(t, logrus.InfoLevel, lgrus.Level)
		assert.Equal(t, len(logrus.AllLevels), len(lgrus.Hooks))

		// checking existence of log file
		logFile := fmt.Sprintf("%s/%s.log", c.FileRedirectPath, c.FileRedirectPrefix)
		info, err := os.Stat(logFile)
		if os.IsNotExist(err) {
			t.Error("log file is not exist")
			return
		}
		assert.True(t, !info.IsDir())

		// logger.Exit method should not block the code
		lgrus.Exit(1)

		// removing test log file
		e := os.Remove(logFile)
		if e != nil {
			t.Errorf("error while removing log file, error: %v", e)
		}
	})
}

func TestDefaultFieldHook_Fire(t *testing.T) {
	h := defaultFieldHook{defaultFields: map[string]interface{}{"defKey": "defVal"}}
	e := &logrus.Entry{Data: map[string]interface{}{}}
	err := h.Fire(e)
	assert.NoError(t, err)
	assert.EqualValues(t, h.defaultFields, e.Data)
}
