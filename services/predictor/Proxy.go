package predictor

import (
	"reflect"
	"returntypes-langserver/common/errors"
)

type Proxy struct {
	// Predicts the return types of the given methods (which are in a "predictable" format, so in the sentence format)
	// The return types are in the same order the method names were sent.
	Predict func(methodsToPredict []string) ([]string, errors.Error) `rpcmethod:"predict" rpcparams:"methodsToPredict"`
	// Trains the predictor and returns the evaluation result if finished.
	Train func(labels, trainingSet, evaluationSet string) (Evaluation, errors.Error) `rpcmethod:"train" rpcparams:"labels,trainingSet,evaluationSet"`
}

type ProxyFacade struct {
	Proxy Proxy `rpcproxy:"true"`
}

// Predicts the return types of the given methods (which are in a "predictable" format, so in the sentence format)
// The return types are in the same order the method names were sent.
func (p *ProxyFacade) Predict(methodsToPredict []string) ([]string, errors.Error) {
	if err := p.validate(p.Proxy.Predict); err != nil {
		return nil, err
	}
	return p.Proxy.Predict(methodsToPredict)
}

// Trains the predictor and returns the evaluation result if finished.
func (p *ProxyFacade) Train(labels, trainingSet, evaluationSet string) (Evaluation, errors.Error) {
	if err := p.validate(p.Proxy.Train); err != nil {
		return Evaluation{}, err
	}
	return p.Proxy.Train(labels, trainingSet, evaluationSet)
}

func (p *ProxyFacade) validate(fn interface{}) errors.Error {
	fnVal := reflect.ValueOf(fn)
	if !fnVal.IsValid() || fnVal.IsZero() {
		return errors.New("RPC Error", "Interface function does not exist")
	}
	return nil
}
