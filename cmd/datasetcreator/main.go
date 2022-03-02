package main

import (
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/processing"
	"returntypes-langserver/processing/typeclasses"
)

func main() {
	err := configuration.Load()
	SetupLogger()
	if err != nil {
		log.FatalError(err)
	}

	LoadTypeClasses()
	StartDatasetCreation()
}

func SetupLogger() {
	log.SetupFileLogging()
	log.SetLoggingToStdout(true)
	log.SetSilentErrorLogging(!configuration.StrictMode())
}

func LoadTypeClasses() {
	log.Info("Load type classes...\n")
	if err := typeclasses.LoadTypeClasses(); err != nil {
		log.FatalError(err)
	}
}

func StartDatasetCreation() {
	processing.ProcessDatasetCreation()
}
