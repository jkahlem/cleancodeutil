package extractor

import (
	"path/filepath"
	"returntypes-langserver/common/code/java"
	"returntypes-langserver/common/code/packagetree"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/common/utils"
)

const ExtractorErrorTitle = "Extractor Error"

// The extractor reads the output from the crawler to "extract" all methods/classes with their resolved canonical names and writes them as output.
type Extractor struct {
	OutputDir string
	tree      packagetree.Tree
	xmlroots  []java.FileContainer
	err       errors.Error
}

func (extractor *Extractor) Err() errors.Error {
	return extractor.err
}

// Reads all extracted code files of the inputDir (= the crawler's xml output) and creates two files:
// - A file containing class declarations with the classes they extends/implements
// - A file containing all method declarations with their return types
//
// The class names/return type names are resolved to their canonical name (as good as possible)
func (extractor *Extractor) Run(inputFiles []string) {
	log.Info("Load extracted java code...\n")
	extractor.createPackageTree(inputFiles)
	log.Info("Start extracting code elements...\n")
	extractor.extract()
}

// Creates a package tree and loads the java elements into it
func (extractor *Extractor) createPackageTree(inputFiles []string) {
	extractor.tree = packagetree.New()
	java.LoadDefaultPackagesToTree(&extractor.tree)
	extractor.loadJavaFilesFromXMLFiles(inputFiles)
}

// Looks for the extracted code files in the input directory, unmarshals them and inserts the files into the package tree
func (extractor *Extractor) loadJavaFilesFromXMLFiles(inputFiles []string) {
	if extractor.err != nil {
		return
	}

	extractor.xmlroots = make([]java.FileContainer, 0, len(inputFiles))

	for index, path := range inputFiles {
		log.Info("[%d/%d] Read entries of %s\n", index+1, len(inputFiles), filepath.Base(path))

		if !utils.FileExists(path) {
			log.Info("[%d/%d] XML file under %s not found.", index+1, len(inputFiles), path)
			continue
		}
		extractor.err = nil

		xmlroot := extractor.loadJavaFilesFromXMLFile(path)
		if extractor.err != nil {
			log.ReportProblemWithError(extractor.err, "Could not load code information for %s", path)
		} else if xmlroot == nil {
			err := errors.New("Error", "No code information available")
			log.ReportProblemWithError(err, "No code information available for %s", path)
		}

		extractor.loadFilesToPackageTree(xmlroot)
		extractor.xmlroots = append(extractor.xmlroots, xmlroot)
	}
	return
}

// Unmarshals the xml file
func (extractor *Extractor) loadJavaFilesFromXMLFile(xmlpath string) java.FileContainer {
	if extractor.err != nil {
		return nil
	}
	xmlroot, err := java.FromXMLFile(xmlpath)
	extractor.err = err
	return xmlroot
}

// Loads the files of the xml file into the package tree
func (extractor *Extractor) loadFilesToPackageTree(xmlroot java.FileContainer) {
	if extractor.err != nil {
		return
	}
	extractor.err = java.LoadFilesToPackageTree(&extractor.tree, xmlroot)
}

// Extracts classes/methods from the java elements
func (extractor *Extractor) extract() {
	if extractor.err != nil {
		return
	}

	allFileCount := 0
	for i := range extractor.xmlroots {
		allFileCount += len(extractor.xmlroots[i].CodeFiles())
	}

	counter := 0
	methodsRecords, classesRecords := make([][]string, 0), make([][]string, 0)
	for i := range extractor.xmlroots {
		for j := range extractor.xmlroots[i].CodeFiles() {
			counter++
			codeFile := extractor.xmlroots[i].CodeFiles()[j]
			log.Info("[%d/%d] Extract code from file: %s\n", counter, allFileCount, codeFile.FilePath)
			visitor := ExtractionVisitor{
				methods:     methodsRecords,
				classes:     classesRecords,
				packageTree: &extractor.tree,
			}
			codeFile.Accept(&visitor)
			methodsRecords, classesRecords = visitor.methods, visitor.classes
		}
	}
	extractor.writeCsvRecords(configuration.MethodsWithReturnTypesOutputPath(), methodsRecords)
	extractor.writeCsvRecords(configuration.ClassHierarchyOutputPath(), classesRecords)
}

func (extractor *Extractor) writeCsvRecords(path string, records [][]string) {
	if extractor.err != nil {
		return
	}
	extractor.err = csv.WriteCsvRecords(path, records)
}
