package crawler

import (
	"reflect"
	"returntypes-langserver/common/errors"
)

type Proxy struct {
	// Gets the content of a code file as a XML object
	GetFileContent func(path string, options Options) (string, errors.Error) `rpcmethod:"getFileContent" rpcparams:"path,options"`
	// Gets the content of all code files in a directory as one xml object
	GetDirectoryContents func(path string, options Options) (string, errors.Error) `rpcmethod:"getDirectoryContents" rpcparams:"path,options"`
}

type ProxyFacade struct {
	Proxy Proxy `rpcproxy:"true"`
}

// Gets the content of a code file as a XML object
func (p *ProxyFacade) GetFileContent(path string, options Options) (string, errors.Error) {
	if err := p.validate(p.Proxy.GetFileContent); err != nil {
		return "", err
	}
	return p.Proxy.GetFileContent(path, options)
}

// Gets the content of all code files in a directory as one xml object
func (p *ProxyFacade) GetDirectoryContents(path string, options Options) (string, errors.Error) {
	if err := p.validate(p.Proxy.GetDirectoryContents); err != nil {
		return "", err
	}
	return p.Proxy.GetDirectoryContents(path, options)
}

func (p *ProxyFacade) validate(fn interface{}) errors.Error {
	fnVal := reflect.ValueOf(fn)
	if !fnVal.IsValid() || fnVal.IsZero() {
		return errors.New("RPC Error", "Interface function does not exist")
	}
	return nil
}
