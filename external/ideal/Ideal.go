package ideal

import (
	"os"
	"path/filepath"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/utils"
)

func AnalyzeFiles(filepath []string) ([]csv.IdealResult, errors.Error) {
	if err := writeFilePaths(filepath); err != nil {
		return nil, err
	} else if err := runIdeal(); err != nil {
		return nil, err
	}
	return loadResultOutput()
}

func AnalyzeSourceCode(sourceCode string) {
	// TODO: Write source code to some temporary file
	//       and then call AnalyzeFiles([]string{temporaryPath})
	//       Remove temporary file afterwards
}

func writeFilePaths(path []string) errors.Error {
	output := "file,type,junit"
	for _, p := range path {
		output += "\n" + p
	}
	if err := os.WriteFile(filepath.Join(configuration.IdealBinaryDir(), "input.csv"), []byte(output), os.ModePerm); err != nil {
		return errors.Wrap(err, "IDEAL Error", "Could not write input")
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
	records, err := csv.ReadRecords(filepath.Join(configuration.IdealBinaryDir(), "IDEAL_Results.csv"))
	if err != nil {
		return nil, errors.Wrap(err, "IDEAL Error", "Could not open output file")
	}
	return csv.UnmarshalIdealResult(records), nil
}
