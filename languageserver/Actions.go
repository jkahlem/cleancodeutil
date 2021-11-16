package languageserver

import "returntypes-langserver/languageserver/lsp"

// Defines an action with an event for messages in the IDE. The event will be fired on interaction
// for example if the user clicks a button in a message.
type Action struct {
	Name  string
	Event func()
}

func NewAction(name string, event func()) Action {
	return Action{
		Name:  name,
		Event: event,
	}
}

func mapActions(actions []Action) []lsp.MessageActionItem {
	destination := make([]lsp.MessageActionItem, 0, len(actions))
	for _, a := range actions {
		destination = append(destination, mapAction(a))
	}
	return destination
}

func mapAction(action Action) lsp.MessageActionItem {
	return lsp.MessageActionItem{
		Title: action.Name,
	}
}
