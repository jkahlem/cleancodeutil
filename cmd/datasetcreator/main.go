package main

import (
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/processing"
	"returntypes-langserver/processing/git"
	"returntypes-langserver/processing/typeclasses"
)

func main() {
	err := configuration.Load()
	SetupLogger()
	if err != nil {
		log.FatalError(err)
	}

	LoadRepositoryList()
	LoadTypeClasses()
	StartDatasetCreation()
}

func SetupLogger() {
	log.SetupFileLogging()
	log.SetLoggingToStdout(true)
	log.SetSilentErrorLogging(!configuration.StrictMode())
}

func LoadRepositoryList() {
	log.Info("Load repository list...\n")
	if err := git.LoadRepositoryList(); err != nil {
		log.FatalError(err)
	}
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
