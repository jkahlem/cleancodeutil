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

func StartProgress(options Options) {
	progressReporterMutex.Lock()
	defer progressReporterMutex.Unlock()

	progressReporter = &ProgressReporter{
		bar:     progressbar.New(0),
		options: options,
	}
}

func FinishProgress() {
	progressReporterMutex.Lock()
	defer progressReporterMutex.Unlock()

	progressReporter.Finish()
	progressReporter = nil
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
	if !p.options.Silent {
		p.bar.Finish()
	}
}
