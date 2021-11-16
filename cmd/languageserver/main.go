package main

import (
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/log"
	"returntypes-langserver/languageserver"
	"returntypes-langserver/processing/typeclasses"
)

func main() {
	err := configuration.Load()
	configuration.SetLangServMode(true)
	SetupLogger()
	if err != nil {
		log.FatalError(err)
	}

	LoadTypeClasses()
	StartLanguageServer()
}

func SetupLogger() {
	log.SetupFileLogging()
	log.SetupRemoteLogging(configuration.LoggerRemotePort())
}

func LoadTypeClasses() {
	log.Info("Load type classes...\n")
	if err := typeclasses.LoadTypeClasses(); err != nil {
		log.FatalError(err)
	}
}

func StartLanguageServer() {
	log.Info("Startup Language Server\n")
	block := make(chan bool, 1)
	languageserver.Startup()
	// the language server is started in a seperated thread, so block the main thread as it is not used anymore
	// (the server will shutdown using os.Exit if it receives such a method call)
	<-block
}
