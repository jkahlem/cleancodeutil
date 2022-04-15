package crawler

import (
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/common/transfer/rpc"
)

// Handles incoming RPC requests/notifications from the crawler.
type Controller struct{}

// Registers the methods available on this application's side.
func (c *Controller) RegisterMethods(register rpc.MethodRegister) {
	register.RegisterMethod("reportProgress", "progress,total,operation", c.ReportProgress)
	register.RegisterMethod("reportError", "message,stacktrace,filepath", c.ReportError)
}

func (c *Controller) ReportProgress(progress, total int, operation string) {
	progressReporterMutex.Lock()
	defer progressReporterMutex.Unlock()

	if progressReporter != nil {
		progressReporter.ReportProgress(progress, total, operation)
	}
}

func (c *Controller) ReportError(message, stacktrace, filePath string) {
	log.Error(errors.New(CrawlerErrorTitle, "An error occured while parsing %s:\n%s\n%s", filePath, message, stacktrace))
}
