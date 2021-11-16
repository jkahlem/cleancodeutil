package crawler

import (
	"returntypes-langserver/common/errors"
	"returntypes-langserver/common/log"
	"returntypes-langserver/common/rpc"
)

// Handles incoming RPC requests/notifications from the crawler.
type Controller struct{}

// Registers the methods available on this application's side.
func (c *Controller) RegisterMethods(register rpc.MethodRegister) {
	register.RegisterMethod("log", "message", c.Log)
	register.RegisterMethod("error", "message", c.Error)
}

// Logs a message.
func (c *Controller) Log(message string) {
	log.Info(message + "\n")
}

// Logs an error message.
func (c *Controller) Error(message string) {
	log.Error(errors.New(CrawlerErrorTitle, message))
}
