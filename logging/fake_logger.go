package logging

// Fake implementation of the ILogger interface to let you turn
// off logging for whatever reason - e.g. running unit tests over & over
type DummyLogger struct {
}

func NewDummyLogger() *DummyLogger {
	return &DummyLogger{}
}

func (dl *DummyLogger) Log(lv LogLevel, msg string) {}

func (dl *DummyLogger) Logf(lv LogLevel, formatted string, args ...any) {}

func (dl *DummyLogger) QuickFmtLog(lv LogLevel, initialText, delim string, args ...any) {}

func (dl *DummyLogger) LogWithCallerInfo(lv LogLevel, initialText string, f func(int) (pc uintptr, file string, line int, ok bool)) {
}
