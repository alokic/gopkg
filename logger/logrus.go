package logger

import (
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"time"
)

type logger struct {
	logrus.Logger
	context logrus.Fields
}

//WriterHook hook.
type WriterHook struct {
	Writer    io.Writer
	LogLevels []logrus.Level
}

//NewLogrus constructor.
func NewLogrus(level logrus.Level, out io.Writer, context logrus.Fields, hooks ...logrus.Hook) Logger {
	lg := &logger{Logger: logrus.Logger{
		Out:   out,
		Level: level,
		Formatter: &logrus.TextFormatter{
			TimestampFormat: time.StampMilli,
		},
		Hooks: make(logrus.LevelHooks),
	},
		context: context}
	setupLog(lg, hooks)
	return lg
}

func (lg *logger) ContextualLogger(context map[string]interface{}) Logger {
	n := new(logger)
	copier.Copy(n, lg)
	n.context = context
	return n
}

func (lg *logger) Debug(args ...interface{}) {
	lg.WithFields(lg.context).Debug(args...)
}

func (lg *logger) Debugf(format string, args ...interface{}) {
	lg.WithFields(lg.context).Debugf(format, args...)
}

func (lg *logger) Info(args ...interface{}) {
	lg.WithFields(lg.context).Info(args...)
}

func (lg *logger) Infof(format string, args ...interface{}) {
	lg.WithFields(lg.context).Infof(format, args...)
}

func (lg *logger) Warn(args ...interface{}) {
	_, f, l, ok := runtime.Caller(1)
	if ok {
		lg.WithFields(lg.context).WithField("caller", fmt.Sprint(f, ":", l)).Warn(args...)
	} else {
		lg.WithFields(lg.context).Warn(args...)
	}
}

func (lg *logger) Warnf(format string, args ...interface{}) {
	_, f, l, ok := runtime.Caller(1)
	if ok {
		lg.WithFields(lg.context).WithField("caller", fmt.Sprint(f, ":", l)).Warnf(format, args...)
	} else {
		lg.WithFields(lg.context).Warnf(format, args...)
	}
}

func (lg *logger) Error(args ...interface{}) {
	_, f, l, ok := runtime.Caller(1)
	if ok {
		lg.WithFields(lg.context).WithField("caller", fmt.Sprint(f, ":", l)).Error(args...)
	} else {
		lg.WithFields(lg.context).Error(args...)
	}
}

func (lg *logger) Errorf(format string, args ...interface{}) {
	_, f, l, ok := runtime.Caller(1)
	if ok {
		lg.WithFields(lg.context).WithField("caller", fmt.Sprint(f, ":", l)).Errorf(format, args...)
	} else {
		lg.WithFields(lg.context).Errorf(format, args...)
	}
}

func (lg *logger) Fatal(args ...interface{}) {
	_, f, l, ok := runtime.Caller(1)
	if ok {
		lg.WithFields(lg.context).WithField("caller", fmt.Sprint(f, ":", l)).Fatal(args...)
	} else {
		lg.WithFields(lg.context).Fatal(args...)
	}
}

func (lg *logger) Fatalf(format string, args ...interface{}) {
	_, f, l, ok := runtime.Caller(1)
	if ok {
		lg.WithFields(lg.context).WithField("caller", fmt.Sprint(f, ":", l)).Fatalf(format, args...)
	} else {
		lg.WithFields(lg.context).Fatalf(format, args...)
	}
}

func (lg *logger) Log(args ...interface{}) error {
	_, f, l, ok := runtime.Caller(1)
	if ok {
		lg.WithFields(lg.context).WithField("caller", fmt.Sprint(f, ":", l)).Print(args...)
	} else {
		lg.WithFields(lg.context).Print(args...)
	}
	return nil
}

// Fire will be called when some logging function is called with current hook
// It will format log entry to string and write it to appropriate writer
func (hook *WriterHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	_, err = hook.Writer.Write([]byte(line))
	return err
}

// Levels define on which log levels this hook would trigger
func (hook *WriterHook) Levels() []logrus.Level {
	return hook.LogLevels
}

// setupLogs adds hooks to send logs to different destinations depending on level
func setupLog(lg *logger, hooks []logrus.Hook) {
	lg.SetOutput(ioutil.Discard) // Send all logs to nowhere by default

	lg.AddHook(&WriterHook{ // Send logs with level higher than warning to stderr
		Writer: os.Stderr,
		LogLevels: []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
		},
	})
	lg.AddHook(&WriterHook{ // Send info and debug logs to stdout
		Writer: os.Stdout,
		LogLevels: []logrus.Level{
			logrus.InfoLevel,
			logrus.DebugLevel,
		},
	})

	for _, h := range hooks {
		if h != nil {
			lg.AddHook(h)
		}
	}
}
