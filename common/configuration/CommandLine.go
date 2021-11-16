package configuration

import "flag"

// Creates flags for command line arguments.
func initCommandLineArguments() {
	flag.StringVar(&loadedConfig.Cloner.RepositoryListPath, "gitinput", "", "the file containing a list of github urls to clone (https-only)")
	flag.StringVar(&loadedConfig.ProjectInputDir, "input", "", "the main output dir containing some processing results and the final dataset")
	flag.StringVar(&loadedConfig.MainOutputDir, "output", "", "the main output dir containing some processing results and the final dataset")
}

// Parses the command line arguments passed to this application and loads them into the configuration.
func loadCommandLineArguments() {
	flag.Parse()
}
