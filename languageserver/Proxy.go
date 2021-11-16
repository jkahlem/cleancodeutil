package languageserver

import (
	"reflect"
	"returntypes-langserver/common/errors"
	"returntypes-langserver/languageserver/lsp"
)

type Proxy struct {
	PublishDiagnostics func(uri lsp.DocumentURI, diagnostics []lsp.Diagnostic, version int) `rpcmethod:"textDocument/publishDiagnostics" rpcparams:"uri,diagnostics,version"`
	ShowMessage        func(msgType lsp.MessageType, message string)                        `rpcmethod:"window/showMessage" rpcparams:"type,message"`
	ShowMessageRequest func(msgType lsp.MessageType, message string,
		actions []lsp.MessageActionItem) (lsp.MessageActionItem, errors.Error) `rpcmethod:"window/showMessageRequest" rpcparams:"type,message,actions"`
	LogMessage         func(msgType lsp.MessageType, message string)                     `rpcmethod:"window/logMessage" rpcparams:"type,message"`
	GetConfiguration   func(items []lsp.ConfigurationItem) ([]interface{}, errors.Error) `rpcmethod:"workspace/configuration" rpcparams:"items"`
	RegisterCapability func([]lsp.Registration) errors.Error                             `rpcmethod:"client/registerCapability" rpcparams:"registrations"`
}

type ProxyFacade struct {
	Proxy Proxy `rpcproxy:""`
}

func (p *ProxyFacade) PublishDiagnostics(uri lsp.DocumentURI, diagnostics []lsp.Diagnostic, version int) {
	if err := p.validate(p.Proxy.PublishDiagnostics); err != nil {
		return
	}
	p.Proxy.PublishDiagnostics(uri, diagnostics, version)
}

func (p *ProxyFacade) ShowMessage(msgType lsp.MessageType, message string) {
	if err := p.validate(p.Proxy.ShowMessage); err != nil {
		return
	}
	p.Proxy.ShowMessage(msgType, message)
}

func (p *ProxyFacade) ShowMessageRequest(msgType lsp.MessageType, message string, actions []lsp.MessageActionItem) (lsp.MessageActionItem, errors.Error) {
	if err := p.validate(p.Proxy.ShowMessageRequest); err != nil {
		return lsp.MessageActionItem{}, err
	}
	return p.Proxy.ShowMessageRequest(msgType, message, actions)
}

func (p *ProxyFacade) LogMessage(msgType lsp.MessageType, message string) {
	if err := p.validate(p.Proxy.ShowMessage); err != nil {
		return
	}
	p.Proxy.LogMessage(msgType, message)
}

func (p *ProxyFacade) GetConfiguration(items []lsp.ConfigurationItem) ([]interface{}, errors.Error) {
	if err := p.validate(p.Proxy.GetConfiguration); err != nil {
		return nil, nil
	}
	return p.Proxy.GetConfiguration(items)
}

func (p *ProxyFacade) RegisterCapability(registrations []lsp.Registration) errors.Error {
	if err := p.validate(p.Proxy.RegisterCapability); err != nil {
		return nil
	}
	return p.Proxy.RegisterCapability(registrations)
}

func (p *ProxyFacade) validate(fn interface{}) errors.Error {
	fnVal := reflect.ValueOf(fn)
	if !fnVal.IsValid() || fnVal.IsZero() {
		return errors.New("RPC Error", "Interface function does not exist")
	}
	return nil
}
