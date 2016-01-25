package util

type Logger interface {
	DebugEnabled() bool
	Infof(string, ...interface{})
	Debugf(string, ...interface{})
}

type goLoggerWrapper struct {
	GoLogger
	debug bool
}

type GoLogger interface {
	Printf(string, ...interface{})
}

func GoLoggerWrapper(goLog GoLogger, debug bool) Logger {
	return &goLoggerWrapper{goLog, debug}
}

func (g *goLoggerWrapper) DebugEnabled() bool {
	return g.debug
}

func (g *goLoggerWrapper) Infof(format string, v ...interface{}) {
	g.GoLogger.Printf(format, v...)
}

func (g *goLoggerWrapper) Debugf(format string, v ...interface{}) {
	if g.debug {
		g.GoLogger.Printf(format, v...)
	}
}

type DebugLogWriter struct {
	prefix string
	logger Logger
}

func NewDebugLogWriter(prefix string, logger Logger) *DebugLogWriter {
	return &DebugLogWriter{prefix, logger}
}

func (d *DebugLogWriter) Write(p []byte) (n int, err error) {
	d.logger.Debugf("%v %v", d.prefix, string(p))
	return len(p), nil
}
