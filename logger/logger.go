package logger

//Logger interface.
type Logger interface {
	ContextualLogger(map[string]interface{}) Logger
	Debug(...interface{})
	Debugf(string, ...interface{})
	Info(...interface{})
	Infof(string, ...interface{})
	Warn(...interface{})
	Warnf(string, ...interface{})
	Error(...interface{})
	Errorf(string, ...interface{})
	Fatal(...interface{})
	Fatalf(string, ...interface{})
	Log(...interface{}) error
}

//DefLogger for logging.
var DefLogger Logger
