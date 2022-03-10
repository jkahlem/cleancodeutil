package predictor

import (
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/common/transfer/messages"
	"returntypes-langserver/common/transfer/rpc"
	"returntypes-langserver/common/transfer/rpc/jsonrpc"
)

var recoverFailedEventHandler []func(rpc.Recoverer)
var connectionErrorEventHandler []func(rpc.Recoverer)

//go:generate go run ../serviceGenerator

func serviceConfiguration() rpc.ServiceConfiguration {
	conn := &PredictorConnection{}
	messager := messages.NewReadWriter(conn)
	messager.AcceptMediaType(jsonrpc.MediaType)
	messager.SetWritingMimeType(jsonrpc.MediaType)
	return rpc.ServiceConfiguration{
		Connection: conn,
		Messager:   messager,
		OnRecoverFailed: func(r rpc.Recoverer) {
			// Call all handlers which are registered with OnRecoverFailed
			for _, fn := range recoverFailedEventHandler {
				if fn != nil {
					fn(r)
				}
			}
		},
		OnConnectionError: func(r rpc.Recoverer) {
			// Call all handlers which are registered with OnConnectionError
			for _, fn := range connectionErrorEventHandler {
				if fn != nil {
					fn(r)
				}
			}
		},
		OnInterfaceCreationError: func(err errors.Error) {
			log.FatalError(err)
		},
		UseMock: configuration.PredictorUseMock(),
	}
}

type Proxy struct {
	// Predicts the return types of the given methods (which are in a "predictable" format, so in the sentence format)
	// The return types are in the same order the method names were sent.
	Predict func(predictionData []string, targetModel SupportedModels) ([]string, errors.Error) `rpcmethod:"predict" rpcparams:"predictionData,targetModel"`
	// Trains the predictor and returns the evaluation result if finished.
	Train func(trainingSet, evaluationSet, additional string, targetModel SupportedModels) (Evaluation, errors.Error) `rpcmethod:"train" rpcparams:"trainingSet,evaluationSet,additional,targetModel"`

	PredictNew func(predictionData []MethodContext, options Options) ([]MethodValues, errors.Error) `rpcmethod:"predict" rpcparams:"predictionData,options"`
	TrainNew   func(trainData []MethodContext, options Options) errors.Error                        `rpcmethod:"train" rpcparams:"trainData,options"`
	Evaluate   func(evaluationData []Method, options Options) (Evaluation, errors.Error)            `rpcmethod:"evaluate" rpcparams:"evaluationData,options"`
}

// Adds a handler for the RecoverFailed event.
func OnRecoverFailed(handler func(rpc.Recoverer)) {
	recoverFailedEventHandler = append(recoverFailedEventHandler, handler)
}

// Adds a handler for the ConnectionError event.
func OnConnectionError(handler func(rpc.Recoverer)) {
	connectionErrorEventHandler = append(connectionErrorEventHandler, handler)
}
