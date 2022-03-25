package languageserver

import (
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/languageserver/lsp"

	"github.com/google/uuid"
)

const (
	ProgressInitialized = "initialized"
	ProgressStarted     = "started"
	ProgressEnded       = "ended"
	ProgressNotCreated  = "invalid"
)

type Progress struct {
	state       string
	message     string
	cancellable bool
	percentage  int
	values      chan interface{}
	token       interface{}
}

func NewProgress(title, message string, token interface{}) *Progress {
	return &Progress{
		state:   ProgressInitialized,
		message: message,
		values:  make(chan interface{}, 10),
		token:   token,
	}
}

// Creates a new progress and starts it already in a separate go routine. If token is nil, the server will
// try to initiate progress reporting (if the client supports this).
func StartProgress(title, message string, token interface{}) *Progress {
	p := NewProgress(title, message, token)
	go p.Start(title)
	return p
}

func (p *Progress) Start(title string) errors.Error {
	if p.state != ProgressInitialized {
		return errors.New("Error", "Expected progress to be in state '%s', actual: '%s'", ProgressInitialized, p.state)
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
	return p.routine()
}

func (p *Progress) Report(message string, percentage int) errors.Error {
	if p.state != ProgressStarted {
		return errors.New("Error", "Expected progress to be in state '%s', actual: '%s'", ProgressStarted, p.state)
	}
	p.percentage = percentage
	p.values <- lsp.WorkDonePrgoressReport{
		Kind:        "report",
		Message:     message,
		Percentage:  percentage,
		Cancellable: p.cancellable,
	}
	return nil
}

func (p *Progress) SetMessage(message string) errors.Error {
	return p.Report(message, p.percentage)
}

func (p *Progress) SetPercentage(percentage int) errors.Error {
	return p.Report(p.message, percentage)
}

func (p *Progress) Close() errors.Error {
	if p.state != ProgressStarted {
		return errors.New("Error", "Expected progress to be in state '%s', actual: '%s'", ProgressStarted, p.state)
	}
	p.state = ProgressEnded
	p.values <- lsp.WorkDonePrgoressEnd{
		Kind:    "end",
		Message: p.message,
	}
	close(p.values)
	return nil
}

func (p *Progress) routine() errors.Error {
	if p.token == nil {
		// Create a new progress if no token specified
		token := uuid.New().String()
		if err := CreateProgress(token); err != nil {
			p.state = ProgressNotCreated
			return errors.Wrap(err, "Error", "Could not create progress object")
		}
		p.token = token
	}
	for {
		value, ok := <-p.values
		if !ok {
			return nil
		}
		remote().Progress(p.token, value)
	}
}

func (p *Progress) State() string {
	return p.state
}
