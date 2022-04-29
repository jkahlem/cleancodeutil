package processing

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"returntypes-langserver/common/code/java"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/jsonschema"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/common/utils"
	"returntypes-langserver/common/utils/counter"
	"returntypes-langserver/processing/dataset"
	"returntypes-langserver/processing/excelOutputter"
	"returntypes-langserver/processing/extractor"
	"returntypes-langserver/processing/git"
	"returntypes-langserver/processing/projects"
)

type Processor struct {
	projects              []projects.Project
	previousProjects      []projects.Project
	isCrawlerFilesUpdated bool
}

func ProcessDatasetCreation() errors.Error {
	processor := Processor{
		projects: projects.GetProjects(),
	}
	if previousProjects, err := LoadPreviousProjectState(); err != nil {
		return err
	} else {
		processor.previousProjects = previousProjects
	}
	processor.ProcessDatasetCreation()
	return SaveProjectState(processor.projects)
}

func LoadPreviousProjectState() ([]projects.Project, errors.Error) {
	path := filepath.Join(configuration.MainOutputDir(), "project-config.json")
	if !utils.FileExists(path) {
		return nil, nil
	}
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "Error", "Could not load previous project state")
	}

	var projectsFile configuration.ProjectConfigurationFile
	if err := jsonschema.UnmarshalJSONStrict(contents, &projectsFile, configuration.ProjectConfigurationFileSchema); err != nil {
		log.Info("The previous projects state file under %s is malformed. Removing the file might fix this issue, but might cause some steps (like crawler steps) to be executed again.\n", path)
		return nil, errors.Wrap(err, "Error", "Could not load previous project state")
	}

	return projects.MapProjects(projectsFile.Projects), nil
}

func SaveProjectState(currentProjects []projects.Project) errors.Error {
	path := filepath.Join(configuration.MainOutputDir(), "project-config.json")
	file, err := utils.CreateFile(path)
	if err != nil {
		return errors.Wrap(err, "Error", "Could not save project state.")
	}

	projectFileStruct := configuration.ProjectConfigurationFile{
		Projects: projects.MapConfigurationProjects(currentProjects),
	}
	if err := json.NewEncoder(file).Encode(projectFileStruct); err != nil {
		return errors.Wrap(err, "Error", "Could not save project state.")
	}
	return nil
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
			hasUpdatedFiles := extractor.PreprocessSourceCodeForProject(project, p.getPreviousProjectStateFor(project))
			if hasUpdatedFiles {
				p.isCrawlerFilesUpdated = true
			}
		}
	}
}

func (p *Processor) getPreviousProjectStateFor(project projects.Project) *projects.Project {
	for _, previousProject := range p.previousProjects {
		if previousProject.Name() == project.Name() {
			return &previousProject
		}
	}
	return nil
}

func (p *Processor) getSymmetricDifference(projectsA, projectsB []projects.Project) []projects.Project {
	projectSet := make(map[string]*projects.Project)
	for _, project := range projectsA {
		projectSet[project.Name()] = &project
	}
	for _, project := range projectsB {
		if _, exists := projectSet[project.Name()]; exists {
			projectSet[project.Name()] = nil
		} else {
			projectSet[project.Name()] = &project
		}
	}

	difference := make([]projects.Project, 0, len(projectSet))
	for _, value := range projectSet {
		if value != nil {
			difference = append(difference, *value)
		}
	}
	return difference
}

// Creates the basic data for dataset creation (which is a list of all methods and the class hierarchy)
func (p *Processor) createBasicData() {
	if p.isExtractionProcessRequired() {
		extractor := extractor.Extractor{}
		extractor.RunOnProjects(p.projects)
		if extractor.Err() != nil {
			log.FatalError(extractor.Err())
		}
		log.Info("Number of failed type resolutions: %d\n", counter.For(java.UnresolvedTypeCounter).GetCount())
		log.Info("Number of failed type resolutions due to imports of external dependencies: %d\n", counter.For(java.DependencyImportCounter).GetCount())
	}
}

func (p *Processor) isExtractionProcessRequired() bool {
	return configuration.ForceExtraction() || p.isDataForDatasetAvailable() || p.isDataForExtractorUpdated()
}

// Returns true if the basic data for dataset creation is available
func (p *Processor) isDataForDatasetAvailable() bool {
	return p.isMethodsWithReturnTypesAvailable() && p.isClassHierarchyAvailable()
}

func (p *Processor) isDataForExtractorUpdated() bool {
	return len(p.getSymmetricDifference(p.projects, p.previousProjects)) > 0 || p.isCrawlerFilesUpdated
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
	return p.trainMethods()
}

func (p *Processor) trainReturnTypes() errors.Error {
	return dataset.Train(configuration.ReturnTypesValidator)
}

func (p *Processor) trainMethods() errors.Error {
	return dataset.Train(configuration.MethodGenerator)
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
