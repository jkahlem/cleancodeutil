package languageserver

import (
	"returntypes-langserver/languageserver/lsp"

	"github.com/google/uuid"
)

const (
	ProgressInitialized = 0
	ProgressStarted     = 1
	ProgressEnded       = 2
	ProgressNotCreated  = 3
)

type Progress struct {
	state       int
	message     string
	cancellable bool
	percentage  int
	values      chan interface{}
	token       interface{}
}

// Starts progress reporting to the client. Token is the workDoneProgress token for the operation. If token is nil, the server will
// try to initiate progress reporting (if the client supports this).
func StartProgress(title, message string, token interface{}) *Progress {
	p := Progress{
		state:   ProgressInitialized,
		message: message,
		values:  make(chan interface{}, 10),
		token:   token,
	}
	p.Start(title)
	return &p
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

func (p *Progress) Close() {
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
	if p.token == nil {
		// Create a new progress if no token specified
		token := uuid.New().String()
		if err := CreateProgress(token); err != nil {
			p.state = ProgressNotCreated
			return
		}
		p.token = token
	}
	for {
		value, ok := <-p.values
		if !ok {
			return
		}
		remote().Progress(p.token, value)
	}
}
