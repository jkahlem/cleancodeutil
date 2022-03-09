package configuration

import (
	"flag"
	"fmt"
)

// Creates flags for command line arguments.
func initCommandLineArguments() {
	flag.StringVar(&loadedConfig.Cloner.OutputDir, "clonedir", "", "the directory where repositories will be cloned to")
	flag.StringVar(&loadedConfig.MainOutputDir, "output", "", "the main output dir containing some processing results and the final dataset")
	flag.BoolVar(&loadedConfig.ForceExtraction, "force", false, "if set, always tries to recollect data from crawler output")
	flag.Func("model", fmt.Sprintf("defines which model type should be used ('%s' or '%s')", MethodGenerator, ReturnTypesValidator), setModelType)
}

func setModelType(str string) error {
	loadedConfig.ModelType = ModelType(str)
	if str != string(ReturnTypesValidator) && str != string(MethodGenerator) {
		return fmt.Errorf("unknown model type: %s", str)
	}
	return nil
}

// Parses the command line arguments passed to this application and loads them into the configuration.
func loadCommandLineArguments() {
	flag.Parse()
}
