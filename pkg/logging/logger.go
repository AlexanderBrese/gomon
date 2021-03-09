package logging

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/AlexanderBrese/gomon/pkg/configuration"
	"github.com/AlexanderBrese/gomon/pkg/utils"

	colorizer "github.com/fatih/color"
)

type logFunc func(string, ...interface{}) (n int, err error)

func newLogFunc(color *colorizer.Color, cfg *configuration.LogConfiguration) logFunc {
	return func(msg string, v ...interface{}) (n int, err error) {
		msg = trimMessage(msg)
		if len(msg) == 0 {
			return
		}
		msg += "\n"
		if cfg.Time {
			msg = addTime(msg)
		}

		return color.Fprintf(colorizer.Output, msg, v...)
	}
}

type RunWriter struct {
	Logger *Logger
}

func (rw *RunWriter) Write(p []byte) (n int, err error) {
	return rw.Logger.appFunc()("%s", string(p))
}

type ErrorWriter struct {
	Logger *Logger
}

func (e *ErrorWriter) Write(p []byte) (n int, err error) {
	return e.Logger.mainFunc()("%s", string(p))
}

type Logger struct {
	config   *configuration.Configuration
	logFuncs map[string]logFunc
	ll       sync.Mutex
}

func NewLogger(cfg *configuration.Configuration) *Logger {
	colors := cfg.Colors()
	logFuncs := make(map[string]logFunc, len(colors))
	for name, color := range colors {
		logFuncs[name] = newLogFunc(color, cfg.Log)
	}
	logFuncs["default"] = defaultLogFunc()
	return &Logger{
		config:   cfg,
		logFuncs: logFuncs,
	}
}

func (l *Logger) BuildLog() (*os.File, error) {
	buildLog, err := l.config.BuildLog()
	if err != nil {
		return nil, err
	}
	return utils.OpenFile(buildLog)
}

func (l *Logger) BuildError(buildError string) error {
	buildLog, err := l.config.BuildLog()
	if err != nil {
		return err
	}
	f, err := utils.CreateFile(buildLog, []byte(buildError))
	if err != nil {
		return err
	}
	return utils.CloseFile(f)
}

func (l *Logger) Main(format string, v ...interface{}) {
	l.log("Main", format, v)
}

func (l *Logger) mainFunc() logFunc {
	return l.getLogFunc("Main")
}

func (l *Logger) Build(format string, v ...interface{}) {
	l.log("Build", format, v)
}

func (l *Logger) Run(format string, v ...interface{}) {
	l.log("Run", format, v)
}

func (l *Logger) Detection(format string, v ...interface{}) {
	l.log("Detection", format, v)
}

func (l *Logger) Sync(format string, v ...interface{}) {
	l.log("Sync", format, v)
}

func (l *Logger) App(format string, v ...interface{}) {
	l.log("App", format, v)
}

func (l *Logger) appFunc() logFunc {
	return l.getLogFunc("App")
}

func (l *Logger) log(level string, format string, v ...interface{}) {
	utils.WithLockAndLog(&l.ll, func() {
		logFunc := l.getLogFunc(level)
		logFunc(format, v...)
	})
}

func (l *Logger) getLogFunc(level string) logFunc {
	rv := reflect.ValueOf(l.config.Log)
	rv = rv.Elem()
	if !rv.FieldByName(level).IsValid() || !rv.FieldByName(level).Bool() {
		return l.logFuncs["default"]
	}
	return l.logFuncs[level]
}

func trimMessage(msg string) string {
	msg = strings.Replace(msg, "\n", "", -1)
	return strings.TrimSpace(msg)
}

func addTime(msg string) string {
	t := time.Now().Format("15:04:05")
	return fmt.Sprintf("[%s] %s", t, msg)
}

func defaultLogFunc() logFunc {
	return newLogFunc(utils.DefaultColor(), configuration.DefaultConfiguration().Log)
}
