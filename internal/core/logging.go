package core

import (
	"fmt"
	"time"
)

// Level is an ordered severity.
// Only entries with level <= the LogSink's configured Level are processed.
// It is a type alias for int, so Go packages can also define them, and interoperate with them.
type Level = int

const (
	LogLevelOff   Level = 0
	LogLevelFatal Level = 1
	LogLevelError Level = 2
	LogLevelWarn  Level = 3
	LogLevelInfo  Level = 4
	LogLevelDebug Level = 5
	LogLevelTrace Level = 6
	LogLevelMAX   Level = LogLevelTrace
)

func LevelToString(l Level) string {
	switch l {
	case LogLevelOff:
		return "off"
	case LogLevelFatal:
		return "fatal"
	case LogLevelError:
		return "error"
	case LogLevelWarn:
		return "warn"
	case LogLevelInfo:
		return "info"
	case LogLevelDebug:
		return "debug"
	case LogLevelTrace:
		return "trace"
	default:
		return "unknown"
	}
}

type LogProvider interface {
	Log(service, instanceID string, level Level, msg string, data any)
}

type TextStreamLogProvider struct {
	writer TextWrite
}

func NewTextStreamLogProvider(writer TextWrite) *TextStreamLogProvider {
	return &TextStreamLogProvider{writer: writer}
}

func NewFileLogProvider(filename string) *TextStreamLogProvider {
	writer, err := NewFileTextWriterAppendOrCreate(filename)
	if err != nil {
		return nil
	}
	return NewTextStreamLogProvider(writer)
}

func NewStdErrLogProvider() *TextStreamLogProvider {
	writer, err := NewStderrTextWriter()
	if err != nil {
		return nil
	}
	return NewTextStreamLogProvider(writer)
}

func NewStdOutLogProvider() *TextStreamLogProvider {
	writer, err := NewStdoutTextWriter()
	if err != nil {
		return nil
	}
	return NewTextStreamLogProvider(writer)
}

const ISOUTCWithMillis = "2006-01-02T15:04:05.000Z"

func (p *TextStreamLogProvider) Log(service, instanceId string, level Level, msg string, data any) {
	line := fmt.Sprintf("%s,%s,%s,%s,%s,%v\n", time.Now().UTC().Format(ISOUTCWithMillis), service, instanceId, LevelToString(level), msg, data)
	p.writer.WriteString(line)
	flushTextWrite(p.writer)
}

// The Outer Logger Type, that provides the Final API interface
// This sits on the core.Session type.
type Logger struct {
	logProvider LogProvider
	service     string
	instanceId  string
	level       Level
}

// Factory; note: no logProvider is set by default, so no logging will be done.
// No default attributes; they are set via an Init call.
func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) Init(service, instanceId string, level Level, logProvider LogProvider) {
	l.service = service
	l.instanceId = instanceId
	l.level = level
	l.logProvider = logProvider
}

func (l *Logger) Log(level Level, msg string, data any) {
	if l.logProvider == nil {
		return
	}
	if level > l.level || level == LogLevelOff {
		return
	}
	l.logProvider.Log(l.service, l.instanceId, level, msg, data)
}

// Set the LogProvider for this Logger; allowed to be nil.
func (l *Logger) SetLogProvider(logProvider LogProvider) {
	l.logProvider = logProvider
}

func (l *Logger) WillLog(level Level) bool {
	return l.logProvider != nil && level <= l.level && level != LogLevelOff
}

func (l *Logger) FinalType() *Type {
	return loggerType
}

func registerLogProviderMembers(t *Type) {
	t.RegisterNativeClassMethod("newStdOutLogProvider", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(NewStdOutLogProvider())
	})
	t.RegisterNativeClassMethod("newStdErrLogProvider", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(NewStdErrLogProvider())
	})
	t.RegisterNativeClassMethod("newFileLogProvider", func(f *RunFrame, ths *TypeObjPair) {
		filename := f.RP.DataStack.Pop().(string)
		f.RP.DataStack.Push(NewFileLogProvider(filename))
	})
}

// CreateRegisterLogLevelType registers LogLevel as a synthetic sealed type whose class
// properties expose log level constants (name → Int value).
func CreateRegisterLogLevelType() {
	t := &Type{
		name: "LogLevel",
		getZero: func() any {
			return Level(0)
		},
		checkCoerceType: func(value any) any {
			switch x := value.(type) {
			case Int:
				return Level(x)
			default:
				panic(fmt.Errorf("LogLevel.checkCoerceType: value %T is not LogLevel", value))
			}
		},
	}
	registerLogLevelClassConstants(t)
	GetTypeRegistrar().RegisterCurrentType(t)
}

func registerLogLevelClassConstants(t *Type) {
	for _, c := range []struct {
		name string
		val  Level
	}{
		{"Off", LogLevelOff},
		{"Fatal", LogLevelFatal},
		{"Error", LogLevelError},
		{"Warn", LogLevelWarn},
		{"Info", LogLevelInfo},
		{"Debug", LogLevelDebug},
		{"Trace", LogLevelTrace},
	} {
		v := Int(c.val)
		name := c.name
		t.registerClassProperty(name, "Int", func(_ Any) Any {
			return v
		}, nil)
	}
}

func CreateRegisterLogProviderType() {
	t := &Type{
		name:            "LogProvider",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[LogProvider]("LogProvider"),
	}
	registerLogProviderMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	logProviderType = t
}

func registerLoggerMembers(t *Type) {
	t.registerProperty("level", "Int",
		func(inst Any) Any { return Int(inst.(*Logger).level) },
		nil,
	)
	t.registerProperty("service", "String",
		func(inst Any) Any { return inst.(*Logger).service },
		func(inst Any, value Any) { inst.(*Logger).service = value.(string) },
	)
	t.registerProperty("instanceId", "String",
		func(inst Any) Any { return inst.(*Logger).instanceId },
		func(inst Any, value Any) { inst.(*Logger).instanceId = value.(string) },
	)

	t.registerNativeMethod("log", func(f *RunFrame, ths *TypeObjPair) {
		level := Level(f.RP.DataStack.Pop().(Int))
		data := f.RP.DataStack.Pop()
		msg := f.RP.DataStack.Pop().(string)
		ths.Obj.(*Logger).Log(level, msg, data)
	})
	t.registerNativeMethod("init", func(f *RunFrame, ths *TypeObjPair) {
		logProvider := f.RP.DataStack.Pop().(LogProvider)
		level := Level(f.RP.DataStack.Pop().(Int))
		instanceId := f.RP.DataStack.Pop().(string)
		service := f.RP.DataStack.Pop().(string)
		ths.Obj.(*Logger).Init(service, instanceId, level, logProvider)
	})
	t.registerNativeMethod("setLevel", func(f *RunFrame, ths *TypeObjPair) {
		level := Level(f.RP.DataStack.Pop().(Int))
		ths.Obj.(*Logger).level = level
	})
	t.registerNativeMethod("willLog", func(f *RunFrame, ths *TypeObjPair) {
		level := Level(f.RP.DataStack.Pop().(Int))
		f.RP.DataStack.Push(ths.Obj.(*Logger).WillLog(level))
	})
	t.registerNativeMethod("setLogProvider", func(f *RunFrame, ths *TypeObjPair) {
		logProvider := f.RP.DataStack.Pop().(LogProvider)
		ths.Obj.(*Logger).SetLogProvider(logProvider)
	})
}

func CreateRegisterLoggerType() {
	t := &Type{
		name:            "Logger",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[Logger]("Logger"),
	}
	registerLoggerMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	loggerType = t
}

var (
	loggerType      *Type
	logProviderType *Type
)

var (
	_ LogProvider = (*TextStreamLogProvider)(nil)
)

func init() {
	CreateRegisterLogLevelType()
	CreateRegisterLogProviderType()
	CreateRegisterLoggerType()
}
