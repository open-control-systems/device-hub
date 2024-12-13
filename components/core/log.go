package core

import (
	"log"
	"os"
)

var (
	// LogInf logs informational events.
	LogInf = log.New(os.Stderr, "inf:", log.LstdFlags)
	// LogWrn logs warning events.
	LogWrn = log.New(os.Stderr, "wrn:", log.LstdFlags)
	// LogErr logs error events.
	LogErr = log.New(os.Stderr, "err:", log.LstdFlags)
)

// SetLogFile setups a log file for all loggers.
func SetLogFile(path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	for _, logger := range []*log.Logger{LogInf, LogWrn, LogErr} {
		logger.SetOutput(file)
		logger.SetFlags(log.LUTC | log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	}

	return nil
}
