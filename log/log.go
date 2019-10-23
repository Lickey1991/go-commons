package log

import (
	"io"
	"os"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"github.com/shiena/ansicolor"
	"github.com/sirupsen/logrus"
)

var (
	//default settings
	settings = Settings{
		Output:       "std",
		Format:       "colored",
		Level:        "info",
		ReportCaller: true,
	}
	lock   sync.Mutex
	logger *logrus.Logger
)

// export Settings
// supporting ini/yaml/json

type LogSettings = Settings

type Settings struct {
	Output       string            `json:"output" yaml:"output" ini:"output"`
	Format       string            `json:"format" yaml:"format" ini:"format"`
	Level        string            `json:"level" yaml:"level" ini:"level"`
	ReportCaller bool              `json:"reportCaller" yaml:"report-caller" ini:"report-caller"`
	FileConfig   *RotateFileConfig `json:"fileConfig" yaml:"file-config" ini:"file-config"`
}

// export GetLogger
func GetLogger(c interface{}) *logrus.Logger {
	if logger != nil {
		return logger
	}
	lock.Lock()
	logger = newLogger(c)
	lock.Unlock()
	return logger
}

func NewLogger(c interface{}) *logrus.Logger {
	return newLogger(c)
}

func IgnoreErrors(errors ...interface{}) {
	if len(errors) > 0 {
		return
	}
}

func newLogger(c interface{}) *logrus.Logger {
	var conf = settings
	if c != nil {
		conf = getConf(c)
	}

	l := logrus.New()
	// for windows no color output
	if windows() && strings.EqualFold(conf.Format, "colored") {
		l.SetOutput(ansicolor.NewAnsiColorWriter(getOutput(conf)))
	} else {
		l.SetOutput(getOutput(conf))
	}
	l.SetFormatter(getFormatter(conf))
	l.SetLevel(getLogLevel(conf))
	l.SetReportCaller(conf.ReportCaller)
	if conf.FileConfig != nil {
		hook, err := NewRotateFileHook(*conf.FileConfig)
		if err == nil {
			l.AddHook(hook)
		}
	}
	return l
}

var conf *Settings

// check all fields of a struct
func getConfig(raw interface{}) {
	if v, ok := raw.(Settings); ok {
		conf = &v
	}
	if v, ok := raw.(*Settings); ok && v != nil {
		conf = v
	}
	getType := reflect.TypeOf(raw)
	getValue := reflect.ValueOf(raw)
	if getType.Kind() == reflect.Struct {
		for i := 0; i < getType.NumField(); i++ {
			value := getValue.Field(i).Interface()
			if reflect.TypeOf(value).Kind() != reflect.Struct {
				continue
			}
			getConfig(value)
		}
	}
}

func getConf(raw interface{}) Settings {
	getConfig(raw)
	if conf == nil {
		return settings
	}
	return *conf
}

func windows() bool {
	return strings.EqualFold(runtime.GOOS, "windows")
}

// get log level, default level info
func getLogLevel(settings Settings) logrus.Level {
	switch strings.ToLower(settings.Level) {
	case "trace":
		return logrus.TraceLevel
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	default:
		return logrus.InfoLevel
	}
}

func getFormatter(c Settings) logrus.Formatter {
	switch c.Format {
	case "colored":
		return &logrus.TextFormatter{
			ForceColors:   true,
			FullTimestamp: true,
		}
	case "text":
		return &logrus.TextFormatter{}
	case "json":
		return &logrus.JSONFormatter{}
	default:
		return &logrus.TextFormatter{
			FullTimestamp: true,
			ForceColors:   true,
		}
	}
}

func getOutput(c Settings) io.Writer {
	switch c.Output {
	case "std":
		return os.Stdout

	default:
		return os.Stdout
	}
}
