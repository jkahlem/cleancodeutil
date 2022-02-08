package rpc

import (
	"returntypes-langserver/common/errors"
	"returntypes-langserver/common/messages"
)

type ServiceConfiguration struct {
	Connection               Connection
	Controller               Controller
	Messager                 messages.Messager
	OnRecoverFailed          func(Recoverer)
	OnConnectionError        func(Recoverer)
	OnInterfaceCreationError func(errors.Error)
	UseMock                  bool
}

func BuildInterfaceFromServiceConfiguration(config ServiceConfiguration, proxyFacade interface{}) (Interface, errors.Error) {
	return CreateInterfaceOnConnection(config.Connection, config.Messager).
		WithController(config.Controller).
		WithProxyFacade(proxyFacade).
		OnConnectionError(config.OnConnectionError).
		OnRecoverFailed(config.OnRecoverFailed).
		Finalize()
}
