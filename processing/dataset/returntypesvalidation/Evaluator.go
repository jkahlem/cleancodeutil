package returntypesvalidation

import (
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/processing/dataset/base"
)

type Evaluator struct{}

func NewEvaluator() base.Evaluator {
	return &Evaluator{}
}

func (e *Evaluator) Evaluate(path string) errors.Error {
	return nil
}

/*
 else {
		log.Info("Evaluation result:\n- Accuracy Score: %g\n- Eval loss: %g\n- F1 Score: %g\n- MCC: %g\n", msg.AccScore, msg.EvalLoss, msg.F1Score, msg.MCC)
		return t.saveEvaluationResult(msg)
	}
*/
