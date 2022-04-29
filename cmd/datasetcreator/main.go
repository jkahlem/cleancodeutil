package main

import (
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/processing"
)

func main() {
	err := configuration.Load()
	SetupLogger()
	if err != nil {
		log.FatalError(err)
	}

	if err := processing.ProcessDatasetCreation(); err != nil {
		log.FatalError(err)
	}
}

func SetupLogger() {
	log.SetupFileLogging()
	log.SetLoggingToStdout(true)
	log.SetSilentErrorLogging(!configuration.StrictMode())
}
