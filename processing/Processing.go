package processing

import (
	"os"

	"returntypes-langserver/common/code/java"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/common/utils/counter"
	"returntypes-langserver/processing/dataset"
	"returntypes-langserver/processing/excelOutputter"
	"returntypes-langserver/processing/extractor"
	"returntypes-langserver/processing/git"
	"returntypes-langserver/processing/projects"
	"returntypes-langserver/processing/statistics"
)

type Processor struct {
	projects []projects.Project
}

func ProcessDatasetCreation() {
	processor := Processor{
		projects: projects.GetProjects(),
	}
	processor.ProcessDatasetCreation()
}

// Executes the dataset creation process
func (p *Processor) ProcessDatasetCreation() {
	// First, clone repositories if needed
	p.clone()
	// Load the java code of each repository and preprocess it using the crawler
	p.preprocessJavaCode()
	// Extract method/classes of all of the repositories and put them into one file for methods and one for classes.
	p.createBasicData()
	// Creates excel outputs for excel output configurations
	p.createExcelOutput()
	// Create a dataset based on the method/class files above.
	p.createDataset()
	// Train the predictor
	p.trainPredictor()
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

func (p *Processor) mapProjectsToRepositoryList(projects []projects.Project) []git.RepositoryDefinition {
	repositories := make([]git.RepositoryDefinition, 0, len(projects))
	for _, project := range projects {
		if len(project.GitUri) > 0 {
			repositories = append(repositories, project.ToRepositoryDefinition())
		}
	}
	return repositories
}

// Preprocess java code using the crawler
func (p *Processor) preprocessJavaCode() {
	if err := os.MkdirAll(configuration.CrawlerOutputDir(), 0777); err != nil {
		log.FatalError(errors.Wrap(err, "Error", "Could not create output directory"))
	} else {
		for _, project := range p.projects {
			extractor.PreprocessSourceCodeForProject(project)
		}
	}
}

// Creates the basic data for dataset creation (which is a list of all methods and the class hierarchy)
func (p *Processor) createBasicData() {
	if !p.isDataForDatasetAvailable() {
		extractor := extractor.Extractor{}
		extractor.RunOnProjects(p.projects)
		if extractor.Err() != nil {
			log.FatalError(extractor.Err())
		}
		log.Info("Number of failed type resolutions: %d\n", counter.For(java.UnresolvedTypeCounter).GetCount())
		log.Info("Number of failed type resolutions due to imports of external dependencies: %d\n", counter.For(java.DependencyImportCounter).GetCount())
	}
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
	if err := excelOutputter.CreateOutput(p.projects); err != nil {
		log.FatalError(err)
	}
}

// Creates a dataset
func (p *Processor) createDataset() {
	if err := dataset.CreateTrainingAndEvaluationSet(configuration.MethodGenerator, configuration.MethodsWithReturnTypesOutputPath(), configuration.ClassHierarchyOutputPath()); err != nil {
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
		log.ReportProblemWithError(errors.Wrap(err, "Training", "Could not train the predictor"), "Could not train the predictor\n")
	} else {
		log.Info("Evaluate...\n")
		if err := dataset.Evaluate(configuration.MethodGenerator); err != nil {
			log.ReportProblemWithError(errors.Wrap(err, "Evaluation", "Could not evaluate datasets"), "Could not evaluate datasets\n")
		}
	}

}

// Executes the training process
func (p *Processor) train() errors.Error {
	/*if err := trainReturnTypes(); err != nil {
		return err
	}*/
	return p.trainMethods() // TODO
}

func (p *Processor) trainReturnTypes() errors.Error {
	return dataset.Train(configuration.ReturnTypesValidator)
}

func (p *Processor) trainMethods() errors.Error {
	return dataset.Train(configuration.MethodGenerator)
}

// Creates statistics for the dataset creation
func (p *Processor) createStatistics() {
	if !configuration.StatisticsSkipCreation() {
		if err := statistics.CreateStatistics(p.projects); err != nil {
			log.ReportProblemWithError(err, "The statistics creation was not successful")
		}
	}
}

// Logs any problems occured during dataset creation
func (p *Processor) logProblems() {
	problems := log.GetProblems()
	if len(problems) > 0 {
		log.Info("During the dataset creation the following problems occured which may have influence on the quality and completeness of the resulting dataset:\n")
		for _, problem := range log.GetProblems() {
			log.Info("- %s\n", problem)
		}
		log.Info("For more information, see the contents of the logfile.\n")
	}
}
