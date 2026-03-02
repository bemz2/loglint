package zap

type Logger struct{}

type SugaredLogger struct{}

func NewProduction(...Option) (*Logger, error) {
	return &Logger{}, nil
}

func (l *Logger) Sync() error {
	return nil
}

func (l *Logger) Sugar() *SugaredLogger {
	return &SugaredLogger{}
}

func (l *Logger) Debug(string, ...Field) {}

func (l *Logger) Info(string, ...Field) {}

func (l *Logger) Warn(string, ...Field) {}

func (l *Logger) Error(string, ...Field) {}

func (s *SugaredLogger) Debugf(string, ...any) {}

func (s *SugaredLogger) Infof(string, ...any) {}

func (s *SugaredLogger) Warnf(string, ...any) {}

func (s *SugaredLogger) Errorf(string, ...any) {}

func (s *SugaredLogger) Debugw(string, ...any) {}

func (s *SugaredLogger) Infow(string, ...any) {}

func (s *SugaredLogger) Warnw(string, ...any) {}

func (s *SugaredLogger) Errorw(string, ...any) {}

type Field struct{}

type Option interface{}
