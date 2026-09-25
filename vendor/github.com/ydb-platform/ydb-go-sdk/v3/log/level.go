package log

import "strings"

type Level int

const (
	TRACE = Level(iota)
	DEBUG
	INFO
	WARN
	ERROR
	FATAL

	QUIET
)

type Mapper interface {
	MapLogLevel(level Level) Level
}

const (
	lblTrace = "TRACE"
	lblDebug = "DEBUG"
	lblInfo  = "INFO"
	lblWarn  = "WARN"
	lblError = "ERROR"
	lblFatal = "FATAL"
	lblQuiet = "QUIET"
)

const (
	colorReset = "\033[0m"

	colorTrace = "\033[38m"
	colorDebug = "\033[37m"
	colorInfo  = "\033[36m"
	colorWarn  = "\033[33m"
	colorError = "\033[31m"
	colorFatal = "\033[41m"
	colorQuiet = colorReset

	colorTraceBold = "\033[47m"
	colorDebugBold = "\033[100m"
	colorInfoBold  = "\033[106m"
	colorWarnBold  = "\u001B[30m\033[103m"
	colorErrorBold = "\033[101m"
	colorFatalBold = "\033[101m"
	colorQuietBold = ""
)

func (l Level) String() string {
	switch l {
	case TRACE:
		return lblTrace
	case DEBUG:
		return lblDebug
	case INFO:
		return lblInfo
	case WARN:
		return lblWarn
	case ERROR:
		return lblError
	case FATAL:
		return lblFatal
	default:
		return lblQuiet
	}
}

func (l Level) BoldColor() string {
	switch l {
	case TRACE:
		return colorTraceBold
	case DEBUG:
		return colorDebugBold
	case INFO:
		return colorInfoBold
	case WARN:
		return colorWarnBold
	case ERROR:
		return colorErrorBold
	case FATAL:
		return colorFatalBold
	default:
		return colorQuietBold
	}
}

func (l Level) Color() string {
	switch l {
	case TRACE:
		return colorTrace
	case DEBUG:
		return colorDebug
	case INFO:
		return colorInfo
	case WARN:
		return colorWarn
	case ERROR:
		return colorError
	case FATAL:
		return colorFatal
	default:
		return colorQuiet
	}
}

func FromString(l string) Level {
	switch strings.ToUpper(l) {
	case lblTrace:
		return TRACE
	case lblDebug:
		return DEBUG
	case lblInfo:
		return INFO
	case lblWarn:
		return WARN
	case lblError:
		return ERROR
	case lblFatal:
		return FATAL
	default:
		return QUIET
	}
}
