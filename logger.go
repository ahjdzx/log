package log

type Namer interface {
	Name() string
}

type Leveler interface {
	Level() LevelType
	SetLevel(lv LevelType)
}

type NamedLeveler interface {
	Namer
	Leveler
}

type MultiHandler interface {
	AddHandler(h Handler)
	RemoveHandler(h Handler)
	Handlers() []Handler
}

// SimpleLogger represents a named logger which is capable of logging with
// multiple handlers and different levels.
//
// Normally this is the logger you should use.
type SimpleLogger interface {
	// Basic
	NamedLeveler

	// multiple handlers
	MultiHandler

	// level APIs
	Debugger
	Printer
	Infoer
	Warner
	Errorer
	Fataler
}

type RPCLogger interface {
	SimpleLogger
	// RPC APIs
	RPCID() string
	RequestID() string
	SetRPCID(rpcID string)
	SetRequestID(requestID string)
}

// Debugger represents a logger with Debug APIs
type Debugger interface {
	Debug(a ...interface{})
	Debugf(format string, a ...interface{})
}

// Printer represents a logger with Print APIs
type Printer interface {
	Print(a ...interface{})
	Println(a ...interface{})
	Printf(f string, a ...interface{})
}

// Infoer represents a logger with Info APIs
type Infoer interface {
	Info(a ...interface{})
	Infof(f string, a ...interface{})
}

// Warner represents a logger with Warn APIs
type Warner interface {
	Warn(a ...interface{})
	Warnf(f string, a ...interface{})
}

// Errorer represents a logger with Error APIs
type Errorer interface {
	Error(a ...interface{})
	Errorf(f string, a ...interface{})
}

// Fataler represents a logger with Fatal APIs
type Fataler interface {
	Fatal(a ...interface{})
	Fatalf(f string, a ...interface{})
}
