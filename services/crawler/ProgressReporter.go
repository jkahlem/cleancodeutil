package crawler

import (
	"returntypes-langserver/common/utils/progressbar"
	"sync"
)

var progressReporter *ProgressReporter
var progressReporterMutex sync.Mutex

type ProgressReporter struct {
	bar     *progressbar.ProgressBar
	options Options
}

func StartProgress(options Options) *ProgressReporter {
	progressReporterMutex.Lock()
	defer progressReporterMutex.Unlock()

	return &ProgressReporter{
		bar:     progressbar.New(0),
		options: options,
	}
}

func (p *ProgressReporter) ReportProgress(progress, total int, operation string) {
	if p.options.Silent {
		return
	}
	p.bar.SetTotal(total).SetCurrent(progress).SetOperation(operation)
	if !p.bar.IsStarted() {
		p.bar.Start()
	}
}

func (p *ProgressReporter) Finish() {
	progressReporterMutex.Lock()
	defer progressReporterMutex.Unlock()

	if !p.options.Silent {
		p.bar.Finish()
	}
	progressReporter = nil
}
