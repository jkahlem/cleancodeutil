package progressbar

import (
	"fmt"

	"github.com/cheggaaa/pb/v3"
)

// Wraps the implementation of a progress bar for the terminal
type ProgressBar struct {
	bar *pb.ProgressBar
}

const Operation = "operation"

const DefaultTemplate = `{{string . "operation"}}{{string . "prefix"}}{{counters . }} {{bar . }} {{percent . }} {{speed . }}{{string . "suffix"}}`

func New(total int) *ProgressBar {
	bar := pb.New(total)
	bar.SetTemplateString(DefaultTemplate)
	return &ProgressBar{
		bar: bar,
	}
}

func StartNew(total int) *ProgressBar {
	return New(total).Start()
}

func (p *ProgressBar) Start() *ProgressBar {
	p.bar.Start()
	return p
}

func (p *ProgressBar) IsStarted() bool {
	return p.bar.IsStarted()
}

func (p *ProgressBar) Current() int {
	return int(p.bar.Current())
}

func (p *ProgressBar) Add(value int) *ProgressBar {
	p.bar.Add(value)
	return p
}

func (p *ProgressBar) SetCurrent(value int) *ProgressBar {
	p.bar.SetCurrent(int64(value))
	return p
}

func (p *ProgressBar) SetTotal(total int) *ProgressBar {
	p.bar.SetTotal(int64(total))
	return p
}

func (p *ProgressBar) AddTotal(total int) *ProgressBar {
	p.bar.AddTotal(int64(total))
	return p
}

func (p *ProgressBar) Total() int {
	return int(p.bar.Total())
}

func (p *ProgressBar) SetOperation(operation string, args ...interface{}) *ProgressBar {
	if operation != "" {
		p.bar.Set(Operation, fmt.Sprintf(operation+"\n", args...))
	} else {
		p.bar.Set(Operation, nil)
	}
	return p
}

func (p *ProgressBar) Finish() *ProgressBar {
	p.SetOperation("")
	p.bar.Finish()
	return p
}
