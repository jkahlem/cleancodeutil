package main

import (
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/languageserver"
)

func main() {
	err := configuration.Load(true)
	SetupLogger()
	if err != nil {
		log.FatalError(err)
	}

	StartLanguageServer()
}

func SetupLogger() {
	log.SetupFileLogging()
	log.SetupRemoteLogging(configuration.LoggerRemotePort())
}

func StartLanguageServer() {
	log.Info("Startup Language Server\n")
	block := make(chan bool, 1)
	languageserver.Startup()
	// the language server is started in a seperated thread, so block the main thread as it is not used anymore
	// (the server will shutdown using os.Exit if it receives such a method call)
	<-block
}
