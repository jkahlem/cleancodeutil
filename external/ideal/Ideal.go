package ideal

import (
	"os"
	"path/filepath"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/utils"
)

var InputWritingError = errors.ErrorId("IDEAL", "Could not write input")
var NotConfiguredError = errors.ErrorId("IDEAL", "Cannot use IDEAL because of missing configurations.")

func AnalyzeFiles(filepath []string) ([]csv.IdealResult, errors.Error) {
	if !IsIdealConfigured() {
		return nil, errors.NewById(NotConfiguredError)
	} else if err := writeFilePaths(filepath); err != nil {
		return nil, err
	} else if err := runIdeal(); err != nil {
		return nil, err
	}
	return loadResultOutput()
}

func AnalyzeSourceCode(sourceCode string) ([]csv.IdealResult, errors.Error) {
	if !IsIdealConfigured() {
		return nil, errors.NewById(NotConfiguredError)
	}
	file, err := os.CreateTemp("", "ideal-input")
	defer os.Remove(file.Name())
	defer file.Close()
	if err != nil {
		return nil, errors.WrapById(err, InputWritingError)
	}
	if _, err := file.Write([]byte(sourceCode)); err != nil {
		return nil, errors.WrapById(err, InputWritingError)
	}
	file.Close()
	return AnalyzeFiles([]string{file.Name()})
}

func writeFilePaths(path []string) errors.Error {
	output := "file,type,junit"
	for _, p := range path {
		output += "\n" + p + ",,"
	}
	if err := os.WriteFile(filepath.Join(configuration.IdealBinaryDir(), "input.csv"), []byte(output), os.ModePerm); err != nil {
		return errors.WrapById(err, InputWritingError)
	}
	return nil
}

func runIdeal() errors.Error {
	p := utils.NewProcess(filepath.Join(configuration.IdealBinaryDir(), "runOnce.cmd"))
	if err := p.Start(); err != nil {
		return err
	} else if err := p.Wait(); err != nil {
		return err
	}
	return nil
}

func loadResultOutput() ([]csv.IdealResult, errors.Error) {
	records, err := csv.NewFileReader(configuration.IdealBinaryDir(), "IDEAL_Results.csv").
		WithSeparator(','). // IDEAL uses comma separator
		SkipFirstLines(1).  // IDEAL results start with a header line
		ReadIdealResultRecords()
	if err != nil {
		return nil, errors.Wrap(err, "IDEAL", "Could not open output file")
	}
	return records, nil
}

func IsIdealConfigured() bool {
	return configuration.IdealBinaryDir() != ""
}
