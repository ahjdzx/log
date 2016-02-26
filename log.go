package log

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"
)

type LevelType int

const (
	NOTSET LevelType = iota
	DEBUG
	INFO
	WARN
	FATA
)

var (
	globalLevel = NOTSET
	logLevel    string
	globalAppID = ""
)

var LevelName = map[LevelType]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	FATA:  "FATA",
}

var LevelColor = map[LevelType]Color{
	DEBUG: Blue,
	INFO:  Green,
	WARN:  Yellow,
	FATA:  Red,
}

var levelFlag = map[string]LevelType{
	"debug": DEBUG,
	"info":  INFO,
	"warn":  WARN,
	"fata":  FATA,
}

type logger struct {
	sync.RWMutex
	wg        sync.WaitGroup
	name      string
	lv        LevelType
	tpl       *template.Template
	handlers  map[Handler]bool
	rpcID     string
	requestID string
}

func New(name string) Logger {
	return NewWithWriter(name, os.Stdout)
}

func NewWithWriter(name string, w io.Writer) Logger {
	l := new(logger)
	l.name = name
	l.lv = NOTSET
	l.handlers = make(map[Handler]bool)
	if w != nil {
		hdr, err := NewStreamHandler(w, defaultTpl)
		if err != nil {
			panic(err)
		}
		l.AddHandler(hdr)
	}
	return l
}

func NewRPCLogger(name string) RPCLogger {
	return New(name).(RPCLogger)
}

func SetGlobalLevel(lv LevelType) {
	globalLevel = lv
}

func GlobalLevel() LevelType {
	return globalLevel
}

func SetGlobalAppID(appID string) {
	globalAppID = appID
}

// AttachFlagSet set some flag, if flagSet is nil, will use flag.CommandLine
func AttachFlagSet(flagSet *flag.FlagSet) {
	if flagSet == nil {
		flagSet = flag.CommandLine
	}
	flagSet.StringVar(&logLevel, "log", "info", "logs at or above this level to the logging output: debug, info, warn, fata")
}

func ParseFlag() error {
	lvl, ok := levelFlag[strings.ToLower(logLevel)]
	if ok {
		globalLevel = lvl
		return nil
	}
	return errors.New("unknown log level")
}

func (l *logger) Name() string {
	l.RLock()
	defer l.RUnlock()
	return l.name
}

func (l *logger) AddHandler(h Handler) {
	l.Lock()
	defer l.Unlock()
	if !l.handlers[h] {
		l.handlers[h] = true
	}
}

func (l *logger) Handlers() []Handler {
	l.RLock()
	defer l.RUnlock()
	var hs = make([]Handler, 0, len(l.handlers))
	for h := range l.handlers {
		hs = append(hs, h)
	}
	return hs
}

func (l *logger) RemoveHandler(h Handler) {
	l.Lock()
	defer l.Unlock()
	delete(l.handlers, h)
}

// Level returns the current level of logger
//
// logger.SetLevel is always authoritative, GlobalLevel is used if SetLevel is
// not called, otherwise defaultLevel is used.
//
// Level() search priority:
// 1. logger's own level (if set)
// 2. GlobalLevel (if set)
// 3. defaultLevel (built-in, usually INFO)
func (l *logger) Level() LevelType {
	l.RLock()
	defer l.RUnlock()
	if l.lv != NOTSET {
		return l.lv
	}
	if globalLevel != NOTSET {
		return globalLevel
	}
	return defaultLevel
}

// SetLevel set the level of logger
//
// SetLevel is always authoritative, See also logger.Level()
func (l *logger) SetLevel(lv LevelType) {
	l.Lock()
	defer l.Unlock()
	l.lv = lv
}

func (l *logger) SetRPCID(rpcID string) {
	l.Lock()
	defer l.Unlock()
	l.rpcID = rpcID
}

func (l *logger) RPCID() string {
	l.RLock()
	defer l.RUnlock()
	return l.rpcID
}

func (l *logger) SetRequestID(requestID string) {
	l.Lock()
	defer l.Unlock()
	l.requestID = requestID
}

func (l *logger) RequestID() string {
	l.RLock()
	defer l.RUnlock()
	return l.requestID
}

func (l *logger) Output(calldepth int, lv LevelType, s string) {
	if lv < l.Level() {
		return
	}
	fileLine := ""
	_, file, line, ok := runtime.Caller(calldepth)
	if !ok {
		file = "???"
		line = 0
	}
	fileLine = file + ":" + strconv.Itoa(line)

	l.RLock()
	r := &Record{
		fileLine:  fileLine,
		name:      l.name,
		now:       time.Now(),
		lv:        lv,
		msg:       s,
		rpcID:     l.rpcID,
		requestID: l.requestID,
		appID:     globalAppID,
	}
	l.RUnlock()

	var wg sync.WaitGroup
	l.Lock()
	defer l.Unlock()
	for h := range l.handlers {
		wg.Add(1)
		go func(h Handler, r *Record) {
			defer wg.Done()
			h.Log(r)
		}(h, r)
	}
	wg.Wait()
}

// Debug APIs
func (l *logger) Debug(a ...interface{}) {
	l.Output(2, DEBUG, fmt.Sprint(a...))
}

func (l *logger) Debugf(format string, a ...interface{}) {
	l.Output(2, DEBUG, fmt.Sprintf(format, a...))
}

// Print APIs output logs with default level
func (l *logger) Print(a ...interface{}) {
	l.Output(2, l.Level(), fmt.Sprint(a...))
}

func (l *logger) Println(a ...interface{}) {
	l.Output(2, l.Level(), fmt.Sprint(a...))
}

func (l *logger) Printf(f string, a ...interface{}) {
	l.Output(2, l.Level(), fmt.Sprintf(f, a...))
}

// Info APIs
func (l *logger) Info(a ...interface{}) {
	l.Output(2, INFO, fmt.Sprint(a...))
}

func (l *logger) Infof(f string, a ...interface{}) {
	l.Output(2, INFO, fmt.Sprintf(f, a...))
}

// Warn APIs
func (l *logger) Warn(a ...interface{}) {
	l.Output(2, WARN, fmt.Sprint(a...))
}

func (l *logger) Warnf(f string, a ...interface{}) {
	l.Output(2, WARN, fmt.Sprintf(f, a...))
}

// Fatal APIs
func (l *logger) Fatal(a ...interface{}) {
	l.Output(2, FATA, fmt.Sprint(a...))
	os.Exit(1)
}

func (l *logger) Fatalf(f string, a ...interface{}) {
	l.Output(2, FATA, fmt.Sprintf(f, a...))
	os.Exit(1)
}
