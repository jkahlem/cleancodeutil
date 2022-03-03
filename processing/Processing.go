package processing

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"returntypes-langserver/common/code/java"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/common/utils"
	"returntypes-langserver/common/utils/counter"
	"returntypes-langserver/processing/dataset"
	"returntypes-langserver/processing/excelOutputter"
	"returntypes-langserver/processing/extractor"
	"returntypes-langserver/processing/git"
	"returntypes-langserver/processing/statistics"
	"returntypes-langserver/services/crawler"
	"returntypes-langserver/services/predictor"
)

type Processor struct {
	projects []Project
}

func ProcessDatasetCreation() {
	processor := Processor{
		projects: GetProjects(),
	}
	processor.ProcessDatasetCreation()
}

// Executes the dataset creation process
func (p *Processor) ProcessDatasetCreation() {
	// First, clone repositories if needed
	p.clone()
	// Load the java code of each repository and summarize it using the crawler
	p.summarizeJavaCode()
	// Extract method/classes of all of the repositories and put them into one file for methods and one for classes.
	p.createBasicData()
	// Creates excel outputs for excel output configurations
	p.createExcelOutput()
	// Create a dataset based on the method/class files above.
	//p.createDataset()
	// Train the predictor
	//p.trainPredictor()
	// Create statistics
	//p.createStatistics()
	// Log any problems occured during creation process
	p.logProblems()
}

// Clone repositories of repository list if not skipped
func (p *Processor) clone() {
	if !configuration.ClonerSkip() {
		log.Info("Start clone process\n")
		repositories := p.mapProjectsToRepositoryList(p.projects)
		if err := git.CloneRepositories(repositories); err != nil {
			log.ReportProblemWithError(err, "The cloning process was not successful")
		}
	}
}

func (p *Processor) mapProjectsToRepositoryList(projects []Project) []git.RepositoryDefinition {
	repositories := make([]git.RepositoryDefinition, 0, len(projects))
	for _, project := range projects {
		if len(project.GitUri) > 0 {
			repositories = append(repositories, project.ToRepositoryDefinition())
		}
	}
	return repositories
}

// Summarize java code using the crawler
func (p *Processor) summarizeJavaCode() {
	if err := os.MkdirAll(configuration.CrawlerOutputDir(), 0777); err != nil {
		log.FatalError(errors.Wrap(err, "Error", "Could not create output directory"))
	} else {
		for _, project := range p.projects {
			p.summarizeJavaCodeForProject(project)
		}
	}
}

// Summarizes the java code for one project
func (p *Processor) summarizeJavaCodeForProject(project Project) {
	// If an output file does already exist, skip summarizing the data for this project.
	if exists, err := p.crawlerOutputFileExists(project); err != nil {
		log.ReportProblemWithError(err, "Could not check if xml output file for %s exists", project.ProjectName())
		return
	} else if exists {
		return
	}

	// Use the crawler to sumamrize the java code structures for a given project into one xml file
	log.Info("Summarize java code for project %s\n", project.ProjectName())
	projectDirPath := project.ExpectedDirectoryPath()
	if !utils.DirExists(projectDirPath) {
		log.ReportProblem("Skip project %s as it does not exist at %s\n", project.ProjectName(), projectDirPath)
		return
	}

	crawlerOptions := crawler.NewOptions().
		Forced(!configuration.StrictMode()).
		WithJavaVersion(project.JavaVersion).
		Build()
	xml, err2 := crawler.GetRawCodeElementsOfDirectory(projectDirPath, crawlerOptions)
	if err2 != nil {
		log.ReportProblemWithError(err2, "Could not create output file for java code files")
	}

	// Write the summarized code structures to an xml file
	file, err := os.Create(p.getXmlPathForProject(project))
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
func (p *Processor) crawlerOutputFileExists(project Project) (bool, errors.Error) {
	_, err := os.Stat(p.getXmlPathForProject(project))
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, errors.Wrap(err, "Error", "Unexpected file error")
}

// Creates the basic data for dataset creation (which is a list of all methods and the class hierarchy)
func (p *Processor) createBasicData() {
	if !p.isDataForDatasetAvailable() {
		extractor := extractor.Extractor{}
		extractor.Run(p.getXmlPathsForProjects(p.projects))
		if extractor.Err() != nil {
			log.FatalError(extractor.Err())
		}
		log.Info("Number of failed type resolutions: %d\n", counter.For(java.UnresolvedTypeCounter).GetCount())
		log.Info("Number of failed type resolutions due to imports of external dependencies: %d\n", counter.For(java.DependencyImportCounter).GetCount())
	}
}

func (p *Processor) getXmlPathsForProjects(projects []Project) []string {
	directories := make([]string, 0, len(projects))
	for _, project := range projects {
		directories = append(directories, p.getXmlPathForProject(project))
	}
	return directories
}

func (p *Processor) getXmlPathForProject(project Project) string {
	return filepath.Join(configuration.CrawlerOutputDir(), project.ProjectName()+".xml")
}

// Returns true if the basic data for dataset creation is available
func (p *Processor) isDataForDatasetAvailable() bool {
	if configuration.ForceExtraction() {
		return false
	} else if ready := p.isMethodsWithReturnTypesAvailable(); !ready {
		return false
	} else if ready := p.isClassHierarchyAvailable(); !ready {
		return false
	}
	return true
}

func (p *Processor) isMethodsWithReturnTypesAvailable() bool {
	if methods, err := os.Stat(configuration.MethodsWithReturnTypesOutputPath()); err != nil {
		return false
	} else if methods.IsDir() {
		return false
	}
	return true
}

func (p *Processor) isClassHierarchyAvailable() bool {
	if classHierarchy, err := os.Stat(configuration.ClassHierarchyOutputPath()); err != nil {
		return false
	} else if classHierarchy.IsDir() {
		return false
	}
	return true
}

func (p *Processor) createExcelOutput() {
	if err := excelOutputter.CreateOutput(); err != nil {
		log.FatalError(err)
	}
}

// Creates a dataset
func (p *Processor) createDataset() {
	if err := dataset.CreateTrainingAndEvaluationSet(configuration.MethodsWithReturnTypesOutputPath(), configuration.ClassHierarchyOutputPath()); err != nil {
		log.FatalError(err)
	}
}

// Trains the predictor with the created dataset if not skipped in configuration
func (p *Processor) trainPredictor() {
	if configuration.PredictorSkipTraining() {
		return
	}
	log.Info("Start training process\n")
	if err := p.train(); err != nil {
		log.ReportProblemWithError(err, "Could not train the predictor")
	}
}

// Executes the training process
func (p *Processor) train() errors.Error {
	/*if err := trainReturnTypes(); err != nil {
		return err
	}*/
	return p.trainMethods()
}

func (p *Processor) trainReturnTypes() errors.Error {
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

func (p *Processor) trainMethods() errors.Error {
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
func (p *Processor) createStatistics() {
	if !configuration.StatisticsSkipCreation() {
		if err := statistics.CreateStatistics(); err != nil {
			log.ReportProblemWithError(err, "The statistics creation was not successful")
		}
	}
}

// Logs any problems occured during dataset creation
func (p *Processor) logProblems() {
	problems := log.GetProblems()
	if !configuration.StrictMode() && len(problems) > 0 {
		log.Info("During the dataset creation the following problems occured which may have influence on the quality and completeness of the resulting dataset:\n")
		for _, problem := range log.GetProblems() {
			log.Info("- %s\n", problem)
		}
		log.Info("For more information, see the contents of the logfile.\n")
	}
}
