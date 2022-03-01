package processing

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"returntypes-langserver/common/code/java"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/common/utils/counter"
	"returntypes-langserver/processing/dataset"
	"returntypes-langserver/processing/excelOutputter"
	"returntypes-langserver/processing/extractor"
	"returntypes-langserver/processing/git"
	"returntypes-langserver/processing/statistics"
	"returntypes-langserver/services/crawler"
	"returntypes-langserver/services/predictor"
)

// Executes the dataset creation process
func ProcessDatasetCreation() {
	// First, clone repositories if needed
	//clone()
	// Load the java code of each repository and summarize it using the crawler
	summarizeJavaCode()
	// Extract method/classes of all of the repositories and put them into one file for methods and one for classes.
	createBasicData()
	// Creates excel outputs for excel output configurations
	createExcelOutput()
	// Create a dataset based on the method/class files above.
	//createDataset()
	// Train the predictor
	//trainPredictor()
	// Create statistics
	//createStatistics()
	// Log any problems occured during creation process
	logProblems()
}

// Clone repositories of repository list if not skipped
func clone() {
	if !configuration.ClonerSkip() {
		log.Info("Start clone process\n")
		if configuration.ClonerRepositoryListPath() == "" {
			log.FatalError(errors.New("Error", "No valid git input file set in configuration or using the '-gitinput' commandline argument."))
		}
		if err := git.CloneRepositories(); err != nil {
			log.ReportProblemWithError(err, "The cloning process was not successful")
		}
	}
}

// Summarize java code using the crawler
func summarizeJavaCode() {
	if files, err := ioutil.ReadDir(configuration.ProjectInputDir()); err != nil {
		log.FatalError(errors.Wrap(err, "Error", "Could not open project input dir"))
	} else if err := os.MkdirAll(configuration.CrawlerOutputDir(), 0777); err != nil {
		log.FatalError(errors.Wrap(err, "Error", "Could not create output directory"))
	} else {
		for _, file := range files {
			if file.IsDir() {
				summarizeJavaCodeForProject(file.Name())
			}
		}
	}
}

// Summarizes the java code for one project
func summarizeJavaCodeForProject(projectDirName string) {
	// If an output file does already exist, skip summarizing the data for this project.
	if exists, err := crawlerOutputFileExists(projectDirName); err != nil {
		log.ReportProblemWithError(err, "Could not check if xml output file for %s exists", projectDirName)
		return
	} else if exists {
		return
	}

	// Use the crawler to sumamrize the java code structures for a given project into one xml file
	log.Info("Summarize java code for project %s\n", projectDirName)
	projectDirPath := filepath.Join(configuration.ProjectInputDir(), projectDirName)
	xml, err2 := crawler.GetRawCodeElementsOfDirectory(projectDirPath, crawler.NewOptions().Forced(!configuration.StrictMode()).Build())
	if err2 != nil {
		log.ReportProblemWithError(err2, "Could not create output file for java code files")
	}

	// Write the summarized code structures to an xml file
	file, err := os.Create(filepath.Join(configuration.CrawlerOutputDir(), projectDirName+".xml"))
	if err != nil {
		wrappedErr := errors.Wrap(err, "Error", "Could not create output file")
		log.ReportProblemWithError(wrappedErr, "Could not create output file for java code files")
	} else if _, err := io.WriteString(file, xml); err != nil {
		wrappedErr := errors.Wrap(err, "Error", "Could not write tooutput file")
		log.ReportProblemWithError(wrappedErr, "Could not write to output file for java code files")
	}
	file.Close()
}

// Returns true if the crawler output file for the given project does exist
func crawlerOutputFileExists(projectDirName string) (bool, errors.Error) {
	_, err := os.Stat(filepath.Join(configuration.CrawlerOutputDir(), projectDirName+".xml"))
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, errors.Wrap(err, "Error", "Unexpected file error")
}

// Creates the basic data for dataset creation (which is a list of all methods and the class hierarchy)
func createBasicData() {
	if !isDataForDatasetAvailable() {
		extractor := extractor.Extractor{}
		extractor.Run(configuration.CrawlerOutputDir())
		if extractor.Err() != nil {
			log.FatalError(extractor.Err())
		}
		log.Info("Number of failed type resolutions: %d\n", counter.For(java.UnresolvedTypeCounter).GetCount())
		log.Info("Number of failed type resolutions due to imports of external dependencies: %d\n", counter.For(java.DependencyImportCounter).GetCount())
	}
}

// Returns true if the basic data for dataset creation is available
func isDataForDatasetAvailable() bool {
	if configuration.ForceExtraction() {
		return false
	} else if ready := isMethodsWithReturnTypesAvailable(); !ready {
		return false
	} else if ready := isClassHierarchyAvailable(); !ready {
		return false
	}
	return true
}

func isMethodsWithReturnTypesAvailable() bool {
	if methods, err := os.Stat(configuration.MethodsWithReturnTypesOutputPath()); err != nil {
		return false
	} else if methods.IsDir() {
		return false
	}
	return true
}

func isClassHierarchyAvailable() bool {
	if classHierarchy, err := os.Stat(configuration.ClassHierarchyOutputPath()); err != nil {
		return false
	} else if classHierarchy.IsDir() {
		return false
	}
	return true
}

func createExcelOutput() {
	if err := excelOutputter.CreateOutput(); err != nil {
		log.FatalError(err)
	}
}

// Creates a dataset
func createDataset() {
	if err := dataset.CreateTrainingAndEvaluationSet(configuration.MethodsWithReturnTypesOutputPath(), configuration.ClassHierarchyOutputPath()); err != nil {
		log.FatalError(err)
	}
}

// Trains the predictor with the created dataset if not skipped in configuration
func trainPredictor() {
	if configuration.PredictorSkipTraining() {
		return
	}
	log.Info("Start training process\n")
	if err := train(); err != nil {
		log.ReportProblemWithError(err, "Could not train the predictor")
	}
}

// Executes the training process
func train() errors.Error {
	/*if err := trainReturnTypes(); err != nil {
		return err
	}*/
	return trainMethods()
}

func trainReturnTypes() errors.Error {
	// Load csv data
	labels, err := csv.ReadRecords(configuration.DatasetLabelsOutputPath())
	if err != nil {
		return err
	}
	trainingSet, err := csv.ReadRecords(configuration.TrainingSetOutputPath())
	if err != nil {
		return err
	}
	evaluationSet, err := csv.ReadRecords(configuration.EvaluationSetOutputPath())
	if err != nil {
		return err
	}

	// Train the predictor
	if msg, err := predictor.TrainReturnTypes(labels, trainingSet, evaluationSet); err != nil {
		return err
	} else {
		// Write the evaluation result in a json file
		if file, err := os.Create(configuration.EvaluationResultOutputPath()); err != nil {
			log.Error(errors.Wrap(err, "Error", "Could not save evaluation result"))
		} else {
			defer file.Close()
			if err := json.NewEncoder(file).Encode(msg); err != nil {
				log.Error(errors.Wrap(err, "Error", "Could not save evaluation result"))
			}
		}
		log.Info("Evaluation result:\n- Accuracy Score: %g\n- Eval loss: %g\n- F1 Score: %g\n- MCC: %g\n", msg.AccScore, msg.EvalLoss, msg.F1Score, msg.MCC)
	}
	return nil
}

func trainMethods() errors.Error {
	// Load csv data
	trainingSet, err := csv.ReadRecords(configuration.MethodsTrainingSetOutputPath())
	if err != nil {
		return err
	}
	/*evaluationSet, err := csv.ReadRecords(configuration.MethodsEvaluationSetOutputPath())
	if err != nil {
		return err
	}*/

	/*formatted := make([][]string, 0, len(trainingSet))
	for _, record := range trainingSet {
		formatted = append(formatted, record[0:1])
	}*/

	// Train the predictor
	if _, err := predictor.TrainMethods(trainingSet[0:40000], nil); err != nil {
		return err
	}
	return nil
}

// Creates statistics for the dataset creation
func createStatistics() {
	if !configuration.StatisticsSkipCreation() {
		if err := statistics.CreateStatistics(); err != nil {
			log.ReportProblemWithError(err, "The statistics creation was not successful")
		}
	}
}

// Logs any problems occured during dataset creation
func logProblems() {
	problems := log.GetProblems()
	if !configuration.StrictMode() && len(problems) > 0 {
		log.Info("During the dataset creation the following problems occured which may have influence on the quality and completeness of the resulting dataset:\n")
		for _, problem := range log.GetProblems() {
			log.Info("- %s\n", problem)
		}
		log.Info("For more information, see the contents of the logfile.\n")
	}
}
