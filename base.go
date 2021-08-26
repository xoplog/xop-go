package xm

type BaseLogger interface {
	SetLevel(Level)
	WantDurable() bool
	StartBuffer() BufferedBase
}

type BufferedBase interface {
	// This is called while holding a lock against other calls to Flush
	Flush()

	Span(
		description string,
		trace Trace,
		parent Trace,
		searchTerms []Field,
		data []Field)

	Prefill([]Field) Prefilled
}

type Prefilled interface {
	Log(level Level, msg string, traceId HexBytes, spanId HexBytes, values []Field)
}

type baseLoggers struct {
	List       []baseLogger
	Removed    []baseLogger
	AnyDurable bool // XXX
}

func (s baseLoggers) CopyWithoutTrace() baseLoggers {
	n := make([]baseLogger, len(s.List))
	for i, bl := range s.List {
		n[i] = baseLogger{
			Name: bl.Name,
			Base: bl.Base,
		}
	}
	return baseLoggers{
		List: n,
	}
}

type baseLogger struct {
	Name      string
	Base      BaseLogger
	Buffered  BufferedBase
	Prefilled Prefilled
}

func WithoutBaseLogger(name string) SeedModifier {
	return func(s *Seed) {
		for i, baseLogger := range s.baseLoggers.List {
			if baseLogger.Name == name {
				s.baseLoggers.Removed = append(s.baseLoggers.Removed, baseLogger)
				if i < len(s.baseLoggers.List)-1 {
					s.baseLoggers.List[i], s.baseLoggers.List[len(s.baseLoggers.List)-1] =
						s.baseLoggers.List[len(s.baseLoggers.List)-1], s.baseLoggers.List[i]
				}
				s.baseLoggers.List = s.baseLoggers.List[:len(s.baseLoggers.List)-1]
				break
			}
		}
	}
}

func WithBaseLogger(name string, base BaseLogger) SeedModifier {
	return func(s *Seed) {
		s.baseLoggers.List = append(s.baseLoggers.List, baseLogger{
			Name: name,
			Base: base,
		})
	}
}

func WithAdditionalPrefill(fields ...Field) SeedModifier {
	return func(s *Seed) {
		s.prefillChanged = true
		s.prefill = append(s.prefill, fields...)
	}
}

func WithOnlyPrefill(fields ...Field) SeedModifier {
	return func(s *Seed) {
		s.prefillChanged = true
		s.prefill = fields
	}
}

func (l *Log) finishBaseLoggerChanges() {
	for i, baseLogger := range l.seed.baseLoggers.List {
		if baseLogger.Buffered == nil {
			baseLogger.Buffered = baseLogger.Base.StartBuffer()
		} else if !l.seed.prefillChanged {
			continue
		}
		baseLogger.Prefilled = baseLogger.Buffered.Prefill(l.seed.prefill)
		l.seed.baseLoggers.List[i] = baseLogger
	}
	for _, baseLogger := range l.seed.baseLoggers.Removed {
		if baseLogger.Buffered != nil {
			baseLogger.Buffered.Flush()
		}
	}
	l.seed.baseLoggers.Removed = nil
	l.seed.prefillChanged = false
}

func (l *Log) BaseLoggers() map[string]BaseLogger {
	m := make(map[string]BaseLogger)
	for _, baseLogger := range l.seed.baseLoggers.List {
		m[baseLogger.Name] = baseLogger.Base
	}
	return m
}
