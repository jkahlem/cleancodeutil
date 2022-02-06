package predictor

import (
	"reflect"
	"returntypes-langserver/common/errors"
)

//go:generate go run ../serviceGenerator

type Proxy struct {
	// Predicts the return types of the given methods (which are in a "predictable" format, so in the sentence format)
	// The return types are in the same order the method names were sent.
	Predict func(predictionData []string, targetModel SupportedModels) ([]string, errors.Error) `rpcmethod:"predict" rpcparams:"predictionData,targetModel"`
	// Trains the predictor and returns the evaluation result if finished.
	Train func(trainingSet, evaluationSet, additional string, targetModel SupportedModels) (Evaluation, errors.Error) `rpcmethod:"train" rpcparams:"trainingSet,evaluationSet,additional,targetModel"`
}

type ProxyFacade struct {
	Proxy Proxy `rpcproxy:"true"`
}

// Predicts the return types of the given methods (which are in a "predictable" format, so in the sentence format)
// The return types are in the same order the method names were sent.
func (p *ProxyFacade) Predict(predictionData []string, targetModel SupportedModels) ([]string, errors.Error) {
	if err := p.validate(p.Proxy.Predict); err != nil {
		return nil, err
	}
	return p.Proxy.Predict(predictionData, targetModel)
}

// Trains the predictor and returns the evaluation result if finished.
func (p *ProxyFacade) Train(trainingSet, evaluationSet, additional string, targetModel SupportedModels) (Evaluation, errors.Error) {
	if err := p.validate(p.Proxy.Train); err != nil {
		return Evaluation{}, err
	}
	return p.Proxy.Train(trainingSet, evaluationSet, additional, targetModel)
}

func (p *ProxyFacade) validate(fn interface{}) errors.Error {
	fnVal := reflect.ValueOf(fn)
	if !fnVal.IsValid() || fnVal.IsZero() {
		return errors.New("RPC Error", "Interface function does not exist")
	}
	return nil
}
