package log

import (
	"errors"
	"log/syslog"
	"net/url"
	"os"
	"strings"

	"github.com/hzwesoft-github/underscore/lang"
	"github.com/sirupsen/logrus"
	syslog_hook "github.com/sirupsen/logrus/hooks/syslog"
)

var (
	logger = logrus.New()

	defaultConfig = &LoggerConfig{
		Module:      "",
		Format:      "json",
		Console:     true,
		File:        false,
		Syslog:      false,
		GlobalLevel: "debug",
	}

	configTemplate = &ResolvedConfig{
		Format: &logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "@at",
				logrus.FieldKeyLevel: "@lvl",
				logrus.FieldKeyMsg:   "@msg",
			}},
		GlobalLevel: logrus.DebugLevel,
	}
)

type LoggerConfig struct {
	Module      string `uci:"module"`
	Format      string `uci:"format"`
	Console     bool   `uci:"console"`
	File        bool   `uci:"file"`
	FilePath    string `uci:"file_path"`
	Syslog      bool   `uci:"syslog"`
	SyslogAddr  string `uci:"syslog_addr"`
	GlobalLevel string `uci:"global_level"`
	SyslogLevel string `uci:"syslog_level"`
}

type ResolvedConfig struct {
	Format      logrus.Formatter
	GlobalLevel logrus.Level
}

type ModuleHook struct {
	Module string
}

func (hook *ModuleHook) Fire(entry *logrus.Entry) error {
	entry.Data["module"] = hook.Module
	return nil
}

func (hook *ModuleHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

type SyslogHookWrapper struct {
	Inst  *syslog_hook.SyslogHook
	Level logrus.Level
}

func (hook *SyslogHookWrapper) Fire(entry *logrus.Entry) error {
	return hook.Inst.Fire(entry)
}

func (hook *SyslogHookWrapper) Levels() []logrus.Level {
	return logrus.AllLevels
}

type ConsoleHook struct {
	Level logrus.Level
}

func (hook *ConsoleHook) Fire(entry *logrus.Entry) (err error) {
	if !logger.IsLevelEnabled(hook.Level) {
		return nil
	}

	var message string
	if message, err = entry.String(); err != nil {
		return err
	}
	if _, err = os.Stdout.WriteString(message); err != nil {
		return err
	}

	return nil
}

func (hook *ConsoleHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func InitLogger(config *LoggerConfig) {
	if config == nil {
		config = defaultConfig
	}

	if !lang.IsBlank(config.Module) {
		logger.AddHook(&ModuleHook{config.Module})
	}

	switch strings.ToUpper(config.Format) {
	case "TEXT":
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	case "JSON":
		fallthrough
	default:
		logger.SetFormatter(configTemplate.Format)
	}

	globalLevel := resolveLogLevel(config.GlobalLevel)
	logger.SetLevel(globalLevel)

	if config.Console {
		logger.AddHook(&ConsoleHook{globalLevel})
	}

	if config.File {
		if lang.IsBlank(config.FilePath) {
			panic(errors.New("ng: log file path is not specific"))
		}

		file, err := os.OpenFile(config.FilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0755)
		if err != nil {
			panic(err)
		}
		logger.SetOutput(file)
	}

	if config.Syslog {
		if lang.IsBlank(config.SyslogAddr) {
			panic(errors.New("ng: syslog addr is not specific"))
		}

		addr, err := url.Parse(config.SyslogAddr)
		if err != nil {
			panic(err)
		}

		hook, err := syslog_hook.NewSyslogHook(addr.Scheme, addr.Host, resolveSyslogPriority(config.SyslogLevel), config.Module)
		if err != nil {
			panic(err)
		}
		logger.AddHook(&SyslogHookWrapper{hook, resolveLogLevel(config.SyslogLevel)})
	}
}

func resolveLogLevel(level string) logrus.Level {
	switch strings.ToLower(level) {
	case "trace":
		return logrus.TraceLevel
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "warning":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	default:
		return logrus.InfoLevel
	}
}

func resolveSyslogPriority(level string) syslog.Priority {
	switch strings.ToLower(level) {
	case "trace":
		return syslog.LOG_DEBUG
	case "debug":
		return syslog.LOG_DEBUG
	case "info":
		return syslog.LOG_INFO
	case "warn":
		return syslog.LOG_WARNING
	case "warning":
		return syslog.LOG_WARNING
	case "error":
		return syslog.LOG_ERR
	case "fatal":
		return syslog.LOG_ALERT
	case "panic":
		return syslog.LOG_EMERG
	default:
		return syslog.LOG_INFO
	}
}

func GetLogger() *logrus.Logger {
	return logger
}
