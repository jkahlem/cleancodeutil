package lsp

type Registration struct {
	Id              string      `json:"id" mapstructure:"id"`
	Method          string      `json:"method" mapstructure:"method"`
	RegisterOptions interface{} `json:"registerOptions" mapstructure:"registerOptions"`
}

type DidChangeConfigurationRegistrationOptions struct {
	Section []string `json:"section,omitempty" mapstructure:"section,omitempty"`
}

func NewRegistration(id, method string, options interface{}) Registration {
	return Registration{
		Id:              id,
		Method:          method,
		RegisterOptions: options,
	}
}
