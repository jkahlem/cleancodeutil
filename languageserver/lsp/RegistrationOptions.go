package lsp

type Registration struct {
	Id              string      `json:"id"`
	Method          string      `json:"method"`
	RegisterOptions interface{} `json:"registerOptions"`
}

type DidChangeConfigurationRegistrationOptions struct {
	Section []string `json:"section,omitempty"`
}

func NewRegistration(id, method string, options interface{}) Registration {
	return Registration{
		Id:              id,
		Method:          method,
		RegisterOptions: options,
	}
}
