package methodgeneration

import (
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/excel"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/services/predictor"

	"github.com/xuri/excelize/v2"
)

type EvaluationResultWriter struct {
	path        string
	file        *excelize.File
	headerStyle *excel.Style
}

func NewResultWriter(path string) (*EvaluationResultWriter, errors.Error) {
	writer := &EvaluationResultWriter{
		file: excelize.NewFile(),
	}
	writer.file.Path = path
	if err := writer.initStyles(); err != nil {
		return nil, err
	}
	return writer, nil
}

func (w *EvaluationResultWriter) initStyles() errors.Error {
	HeaderStyle := excel.Style{
		Bold:            true,
		BackgroundColor: "#ABABAB",
	}
	if _, err := HeaderStyle.ToExcelStyle(w.file); err != nil {
		return err
	}
	w.headerStyle = &HeaderStyle
	return nil
}

// Writes the examples to the result. exampleContexts are the configured examples as predictor.MethodContext objects.
// The generatedOutputs are the predictor.MethodValues which were generated by the predictor.
// len(exampleContexts) must equal len(generatedOutputs), so generatedOutputs[i] is the list of predicted values (suggestions) for
// the example at exampleContexts[i].
func (w *EvaluationResultWriter) WriteExamples(exampleDefinitions []configuration.MethodExample, exampleContexts []predictor.MethodContext,
	generatedOutputs [][]predictor.MethodValues) errors.Error {
	if err := w.check(); err != nil {
		return err
	} else if len(exampleContexts) != len(generatedOutputs) || len(exampleDefinitions) != len(exampleContexts) {
		return errors.New("Evaluation", "Could not write example output: The amount of generated values does not match the amount of examples.")
	}

	sheet := "Examples"
	w.file.NewSheet(sheet)
	w.file.SetColWidth(sheet, excel.GetColumnIdentifier(0), excel.GetColumnIdentifier(0), 50)
	w.file.SetColWidth(sheet, excel.GetColumnIdentifier(2), excel.GetColumnIdentifier(2), 80)

	cursor := excel.NewCursor(w.file, "Examples")
	cursor.SetStyle(w.headerStyle.Id())
	cursor.WriteRowValues("Input", "Label", "Generated outputs")
	cursor.SetStyle(0)
	cursor.Move(0, 1)
	for i, example := range exampleContexts {
		cursor.WriteRowValues(example)
		cursor.Move(1, 0)
		cursor.WriteRowValues(exampleDefinitions[i].Label)
		cursor.Move(1, 0)
		for _, generatedValues := range generatedOutputs[i] {
			cursor.WriteRowValues(CreateMethodDefinition(example, generatedValues))
			cursor.Move(0, 1)
		}
		cursor.Move(-2, 0)
	}
	return cursor.Error()
}

func (w *EvaluationResultWriter) WriteMethods(methods []Method) errors.Error {
	if err := w.check(); err != nil {
		return err
	}

	i := 0
	return excel.Stream().FromFunc(func() []string {
		if i >= len(methods) {
			return nil
		}
		record := w.toMethodRecord(methods[i])
		i++
		return record
	}).WithColumnsFromStruct(MethodRecordLayout{}).
		ToSheet(w.file, "Generated methods")
}

func (w *EvaluationResultWriter) toMethodRecord(method Method) []string {
	return []string{method.Name, method.ExpectedDefinition.String(), method.GeneratedDefinition.String()}
}

func (w *EvaluationResultWriter) WriteScores(evalset *EvaluationSet) errors.Error {
	if err := w.check(); err != nil {
		return err
	} else if evalset == nil {
		return errors.New("Evaluation", "Could not create evaluation result output")
	}

	if err := w.writeScoreSheetForSet(evalset); err != nil {
		return err
	}

	for _, set := range evalset.Subsets {
		if err := w.WriteScores(&set); err != nil {
			return err
		}
	}
	return nil
}

func (w *EvaluationResultWriter) writeScoreSheetForSet(evalset *EvaluationSet) errors.Error {
	if len(evalset.Rater) == 0 {
		return nil
	}

	sheet := "Set - " + evalset.Name
	w.file.NewSheet(sheet)
	w.file.SetColWidth(sheet, excel.GetColumnIdentifier(0), excel.GetColumnIdentifier(0), 50)
	w.file.SetColWidth(sheet, excel.GetColumnIdentifier(1), excel.GetColumnIdentifier(1), 50)
	cursor := excel.NewCursor(w.file, sheet)

	for _, rater := range evalset.Rater {
		cursor.SetStyle(w.headerStyle.Id())
		cursor.WriteRowValues("Rating method:", rater.Name())
		cursor.SetStyle(0)
		cursor.Move(0, 1)
		result := rater.Result()
		cursor.WriteValues(result)
	}

	if err := cursor.Error(); err != nil {
		return err
	}
	return nil
}

func (w *EvaluationResultWriter) check() errors.Error {
	if w.file == nil {
		return errors.New("Evaluation", "The excel output file does not exist")
	} else if w.headerStyle == nil {
		return errors.New("Evaluation", "The result writer was not initialized correctly")
	}
	return nil
}

func (w *EvaluationResultWriter) Close() errors.Error {
	if w.file == nil {
		return nil
	}
	// Delete the default sheet - this needs to be done at the end if there are other sheets.
	w.file.DeleteSheet("Sheet1")
	return errors.Wrap(excel.SaveFile(w.file), "Evaluation", "Could not save output file")
}

type MethodRecordLayout struct {
	Name                string `excel:"Method Name,width=25"`
	ExpectedDefinition  string `excel:"Expected Definition,width=80"`
	GeneratedDefinition string `excel:"Generated Definition,width=80"`
}
