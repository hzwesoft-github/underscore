package log

import (
	"testing"
)

func TestDefaultLog(t *testing.T) {
	InitLogger(nil)

	logger := GetLogger()

	logger.Traceln("this is below default level")
	logger.Debugln("this is below default level")
	logger.Infoln("Info level")
	logger.Errorln("Error level")
}

// func TestCustomLog(t *testing.T) {
// 	config := &LoggerConfig{
// 		Module:      "Test",
// 		Format:      "json",
// 		Console:     true,
// 		File:        true,
// 		FilePath:    "log_test.log",
// 		Syslog:      true,
// 		SyslogAddr:  "udp://192.168.247.143:514",
// 		GlobalLevel: "info",
// 		SyslogLevel: "debug",
// 	}

// 	InitLogger(config)

// 	logger := GetLogger()

// 	logger.Traceln("this is below default level")
// 	logger.Debugln("this is below default level")
// 	logger.Warnln("Warn level")
// 	logger.Infoln("Info level")
// 	logger.Errorln("Error level")

// }
