package languageserver

import "returntypes-langserver/languageserver/lsp"

const (
	ProgressInitialized = 0
	ProgressStarted     = 1
	ProgressEnded       = 2
)

type Progress struct {
	state       int
	message     string
	cancellable bool
	percentage  int
	values      chan interface{}
	token       interface{}
}

func StartProgress(title, message string, token interface{}) Progress {
	p := Progress{
		state:   ProgressInitialized,
		message: message,
		values:  make(chan interface{}),
		token:   token,
	}
	p.Start(title)
	return p
}

func (p *Progress) Start(title string) {
	if p.state != ProgressInitialized {
		return
	}
	p.state = ProgressStarted
	p.values <- lsp.WorkDoneProgressBegin{
		Title: title,
		WorkDonePrgoressReport: lsp.WorkDonePrgoressReport{
			Kind:        "begin",
			Message:     p.message,
			Cancellable: p.cancellable,
			Percentage:  0,
		},
	}
	go p.routine()
}

func (p *Progress) Report(message string, percentage int) {
	if p.state != ProgressStarted {
		return
	}
	p.percentage = percentage
	p.values <- lsp.WorkDonePrgoressReport{
		Kind:        "report",
		Message:     message,
		Percentage:  percentage,
		Cancellable: p.cancellable,
	}
}

func (p *Progress) SetMessage(message string) {
	if p.state != ProgressStarted {
		return
	}
	p.Report(message, p.percentage)
}

func (p *Progress) SetPercentage(percentage int) {
	if p.state != ProgressStarted {
		return
	}
	p.Report(p.message, percentage)
}

func (p *Progress) Finish() {
	if p.state != ProgressStarted {
		return
	}
	p.state = ProgressEnded
	p.values <- lsp.WorkDonePrgoressEnd{
		Kind:    "end",
		Message: p.message,
	}
	close(p.values)
}

func (p *Progress) routine() {
	for {
		value, ok := <-p.values
		if !ok {
			return
		}
		remote().Progress(p.token, value)
	}
}
