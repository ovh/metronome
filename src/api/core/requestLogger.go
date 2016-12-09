package core

import (
	log "github.com/Sirupsen/logrus"
)

type RequestLogger struct {
	LogType string
	Level   log.Level
}

func (rl RequestLogger) Println(v ...interface{}) {
	lg := log.WithField("type", rl.LogType)
	l := lg.Infoln

	switch rl.Level {
	case log.PanicLevel:
		l = lg.Panicln
	case log.FatalLevel:
		l = lg.Fatalln
	case log.ErrorLevel:
		l = lg.Errorln
	case log.WarnLevel:
		l = lg.Warnln
	case log.InfoLevel:
		l = lg.Infoln
	case log.DebugLevel:
		l = lg.Debugln
	}

	l(v...)
}

func (rl RequestLogger) Printf(format string, v ...interface{}) {
	lg := log.WithField("type", rl.LogType)
	l := lg.Infof

	switch rl.Level {
	case log.PanicLevel:
		l = lg.Panicf
	case log.FatalLevel:
		l = lg.Fatalf
	case log.ErrorLevel:
		l = lg.Errorf
	case log.WarnLevel:
		l = lg.Warnf
	case log.InfoLevel:
		l = lg.Infof
	case log.DebugLevel:
		l = lg.Debugf
	}

	l(format, v...)
}
