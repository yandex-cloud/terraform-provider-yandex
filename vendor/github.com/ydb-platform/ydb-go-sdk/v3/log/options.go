package log

type Option interface {
	applyHolderOption(l *wrapper)
}

type coloringSimpleOption bool

func (coloring coloringSimpleOption) applySimpleOption(l *defaultLogger) {
	l.coloring = bool(coloring)
}

func WithColoring() simpleLoggerOption {
	return coloringSimpleOption(true)
}

type minLevelSimpleOption Level

func (minLevel minLevelSimpleOption) applySimpleOption(l *defaultLogger) {
	l.minLevel = Level(minLevel)
}

func WithMinLevel(level Level) simpleLoggerOption {
	return minLevelSimpleOption(level)
}

type logQueryOption bool

func (logQuery logQueryOption) applySimpleOption(l *defaultLogger) {
	l.logQuery = bool(logQuery)
}

func (logQuery logQueryOption) applyHolderOption(l *wrapper) {
	l.logQuery = bool(logQuery)
}

func WithLogQuery() logQueryOption {
	return true
}
