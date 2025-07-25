package languageserver

import (
	"encoding/json"
	"fmt"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/common/transfer/rpc"
	"returntypes-langserver/common/utils"
	"returntypes-langserver/languageserver/diagnostics"
	"returntypes-langserver/languageserver/lsp"
	"returntypes-langserver/languageserver/workspace"
	"returntypes-langserver/processing/dataset/methodgeneration"
	"returntypes-langserver/services/predictor"
	"strings"
)

const (
	MethodGeneratorConfigSection = "methodGenerator"
)

type languageServer struct {
	configuration      ServerConfiguration
	workspaces         workspace.Container
	diagnosticsCreator *diagnostics.Creator
	count              int
	predictorRecoverer rpc.Recoverer
} // @ServiceGenerator:ServiceDefinition

func (ls *languageServer) Configuration() *ServerConfiguration {
	return &ls.configuration
}

func (ls *languageServer) setClientCapabilities(clientCapabilities lsp.ClientCapabilities) {
	ls.configuration.clientCapabilities = clientCapabilities
}

// Starts the language server.
func (ls *languageServer) Startup() {
	// creates an interface with a controller listening to the stdio if it not exist already
	getInterface()

	// Register events for predictor connection errors
	predictor.OnConnectionError(func(r rpc.Recoverer) {
		ls.ShowMessage(lsp.MessageError, "Could not connect to the predictor due to malformed configuration.")
		ls.predictorRecoverer = r
	})
	predictor.OnRecoverFailed(func(r rpc.Recoverer) {
		ls.ShowMessageRequest(lsp.MessageError, "Could not connect to the predictor.", []Action{NewAction("Reconnect", func() {
			r.Recover()
		})})
		ls.predictorRecoverer = r
	})
}

// Create virtual workspaces using the given workspace folders.
func (ls *languageServer) createVirtualWorkspaces(workspaces []lsp.WorkspaceFolder) {
	ls.configuration.workspaces = workspaces
	for _, workspace := range workspaces {
		if err := ls.createVirtualWorkspace(workspace); err != nil {
			log.Error(err)
		}
	}
}

// Create a virtual workspace using the given workspace folder.
func (ls *languageServer) createVirtualWorkspace(workspace lsp.WorkspaceFolder) errors.Error {
	path, err := lsp.DocumentURIToFilePath(workspace.URI)
	ls.log("Create virtual workspace for %s", path)
	if err != nil {
		return err
	}
	ws := ls.workspaces.CreateWorkspace(path)
	return ws.Load()
}

// Adds a file on the given path into the virtual workspace if it does not exist there already.
func (ls *languageServer) AddFileIfNotExists(path, text string) {
	for _, ws := range ls.workspaces.List() {
		if ws.IsFileBelongingToWorkspace(path) && ws.FileSystem.GetFile(path) == nil {
			if err := ws.AddFile(path); err != nil {
				log.Error(err)
				return
			} else if err := ls.refreshDiagnosticsForFile(ws, ws.FileSystem.GetFile(path)); err != nil {
				log.Error(err)
				return
			}
		}
		if file := ws.FileSystem.GetFile(path); file != nil {
			file.Document().SetText(text)
		}
	}
}

// Reloads the file on the given path in all virtual workspaces containing it.
func (ls *languageServer) ReloadFile(path string) {
	for _, ws := range ls.workspaces.List() {
		if ws.IsFileBelongingToWorkspace(path) {
			if err := ws.ReloadFile(path); err != nil {
				log.Error(err)
			}
		}
	}
}

// Renames the file on the given path in all virtual workspaces containing it.
func (ls *languageServer) RenameFile(oldPath, newPath string) {
	for _, ws := range ls.workspaces.List() {
		if ws.IsFileBelongingToWorkspace(oldPath) {
			ws.RenameFile(oldPath, newPath)
		}
	}
}

// Deletes the file on the given path in all virtual workspaces containing it.
func (ls *languageServer) DeleteFile(path string) {
	for _, ws := range ls.workspaces.List() {
		if ws.IsFileBelongingToWorkspace(path) {
			ws.FileSystem.RemoveFile(path)
		}
	}
}

// Returns the file for the given path if it exists
func (ls *languageServer) GetFile(path string) *workspace.FileWrapper {
	for _, ws := range ls.workspaces.List() {
		if ws.IsFileBelongingToWorkspace(path) {
			return ws.FileSystem.GetFile(path)
		}
	}
	return nil
}

// Updates the diagnostics of the given file in all workspaces containing it.
func (ls *languageServer) UpdateDiagnostics(path string, changes []lsp.TextDocumentContentChangeEvent) {
	if !ls.IsReturntypeValidationActive() {
		return
	}
	for _, ws := range ls.workspaces.List() {
		if ws.IsFileBelongingToWorkspace(path) {
			file := ws.FileSystem.GetFile(path)
			updated := false
			for _, change := range changes {
				if file.Diagnostics().UpdatePositions(change) {
					updated = true
				}
			}
			if updated {
				ls.PublishDiagnostics(path, diagnostics.MapExpectedReturnTypeDiagnostics(file.Diagnostics().Diagnostics()), file.Diagnostics().Version())
			}
		}
	}
}

// Updates the diagnostics of the given file in all workspaces containing it.
func (ls *languageServer) UpdateDocuments(path string, changes []lsp.TextDocumentContentChangeEvent) {
	if !ls.IsMethodGenerationActive() {
		return
	}
	for _, ws := range ls.workspaces.List() {
		if ws.IsFileBelongingToWorkspace(path) {
			file := ws.FileSystem.GetFile(path)
			file.Document().ApplyChanges(changes)
		}
	}
}

// Refreshes the diagnostics for all files.
func (ls *languageServer) RefreshDiagnosticsForAllFiles() {
	for _, ws := range ls.workspaces.List() {
		ls.refreshDiagnosticsForAllFilesInWorkspace(ws)
	}
}

// Refrehses the diagnostics for all files in a virtual workspace.
func (ls *languageServer) refreshDiagnosticsForAllFilesInWorkspace(ws *workspace.Workspace) {
	for _, file := range ws.FileSystem.Files() {
		if err := ls.refreshDiagnosticsForFile(ws, file); err != nil {
			log.Error(err)
		}
	}
}

// Refreshes the diagnostics for a given file.
func (ls *languageServer) RefreshDiagnosticsForFile(path string) {
	for _, ws := range ls.workspaces.List() {
		if ws.IsFileBelongingToWorkspace(path) {
			if err := ls.refreshDiagnosticsForFile(ws, ws.FileSystem.GetFile(path)); err != nil {
				log.Error(err)
			}
		}
	}
}

// Refreshes the diagnostics for a file in a virtual workspace.
func (ls *languageServer) refreshDiagnosticsForFile(ws *workspace.Workspace, file *workspace.FileWrapper) errors.Error {
	if !ls.IsReturntypeValidationActive() {
		return nil
	} else if file == nil {
		return errors.New("Error", "No file given to refresh")
	}

	// create diagnsotics for file
	creator := ls.getDiagnosticsCreator()
	if creator == nil {
		return errors.New("Error", "Creator does not exist")
	}
	if d, err := creator.CreateDiagnosticsForFile(file.File(), ws.FileSystem.PackageTree()); err != nil {
		return err
	} else {
		file.Diagnostics().SetDiagnostics(d)
		ls.PublishDiagnostics(file.Path(), diagnostics.MapExpectedReturnTypeDiagnostics(d), file.Diagnostics().Version())
	}
	return nil
}

// Returns a diagnostics creator.
func (ls *languageServer) getDiagnosticsCreator() *diagnostics.Creator {
	if ls.diagnosticsCreator == nil {
		ls.diagnosticsCreator = &diagnostics.Creator{}
	}
	return ls.diagnosticsCreator
}

// Publishes diagnostics to the client.
func (ls *languageServer) PublishDiagnostics(path string, diagnostics []lsp.Diagnostic, version int) {
	if ls.isClientSupportingDiagnostics() {
		if ls.isClientSupportingDiagnosticVersions() {
			remote().PublishDiagnostics(lsp.FilePathToDocumentURI(path), diagnostics, version)
		} else {
			remote().PublishDiagnostics(lsp.FilePathToDocumentURI(path), diagnostics, 0)
		}
	}
}

// Returns true if the client supports versioning of diagnostics.
func (ls *languageServer) isClientSupportingDiagnosticVersions() bool {
	return ls.isClientSupportingDiagnostics() && ls.configuration.PublishDiagnosticsClientCapabilities().VersionSupport
}

// Returns true if the client supports the textDocument/publishDiagnostics method.
func (ls *languageServer) isClientSupportingDiagnostics() bool {
	return ls.configuration.PublishDiagnosticsClientCapabilities() != nil
}

// Shows a message to the user in the IDE.
func (ls *languageServer) ShowMessage(msgType lsp.MessageType, message string) {
	remote().ShowMessage(msgType, message)
}

// Makes a request to the user with possible actions (appears for example as buttons the user can click)
func (ls *languageServer) ShowMessageRequest(msgType lsp.MessageType, message string, actions []Action) {
	go func() {
		if clickedAction, err := remote().ShowMessageRequest(msgType, message, mapActions(actions)); err == nil {
			for _, a := range actions {
				if a.Name == clickedAction.Title {
					a.Event()
				}
			}
		}
	}()
}

// Logs a message to the IDE.
func (ls *languageServer) LogMessage(msgType lsp.MessageType, message string) {
	remote().LogMessage(msgType, message)
}

// Recovers the predictor connection.
func (ls *languageServer) RecoverPredictor() {
	if ls.predictorRecoverer != nil {
		ls.predictorRecoverer.Recover()
	}
}

// Loads the extension configuration of the IDE.
func (ls *languageServer) LoadConfiguration() chan errors.Error {
	promise := make(chan errors.Error)
	go func() {
		promise <- ls.loadConfiguration(MethodGeneratorConfigSection)
	}()
	return promise
}

// Loads the extension configuration of the IDE.
func (ls *languageServer) loadConfiguration(items ...string) errors.Error {
	if !ls.configuration.ConfigurationClientCapabilities() {
		return nil
	}
	results, err := remote().GetConfiguration(lsp.MapConfigurationItems(items...))
	log.Info("%v\n", results)
	for i, config := range results {
		log.Info("%s\n", items[i])
		if items[i] == MethodGeneratorConfigSection {
			if asJson, err := json.Marshal(config); err != nil {
				log.Info("%s\n", string(asJson))
				configuration.LoadConfigFromJsonString(string(asJson))
			}
		}
	}
	return err
}

// Registers workspace/didChangeConfiguration for the extension's configuration sections.
// This capability needs to be registered explicitly, otherwise there will be no notifications.
func (ls *languageServer) RegisterDidChangeWorkspaceCapability() chan errors.Error {
	return ls.RegisterCapability(lsp.NewRegistration(utils.NewUuid(), lsp.MethodWorkspace_DidChangeConfiguration, lsp.DidChangeConfigurationRegistrationOptions{
		Section: []string{MethodGeneratorConfigSection},
	}))
}

// Registers a capability.
func (ls *languageServer) RegisterCapability(registrations ...lsp.Registration) chan errors.Error {
	promise := make(chan errors.Error)
	go func() {
		promise <- remote().RegisterCapability(registrations)
	}()
	return promise
}

// Creates a completion item
func (ls *languageServer) CompleteMethodDefinition(method Method, doc *workspace.Document) ([]lsp.CompletionItem, errors.Error) {
	if doc == nil {
		return nil, nil
	}

	suggestions, err := ls.generateParameterLists(method)
	if err != nil || len(suggestions) == 0 {
		return nil, err
	}

	items := make([]lsp.CompletionItem, len(suggestions[0]))
	for i, suggestion := range suggestions[0] {
		// convert output to completion item & return it
		parameterList := methodgeneration.ConcatParametersToList(suggestion.Parameters)
		parameterListTextEdit := ls.createTextEdit(parameterList, lsp.Range{
			Start: doc.ToPosition(method.RoundBraces.Range.Start + 1),
			End:   doc.ToPosition(method.RoundBraces.Range.End - 1),
		})
		if !method.Type.IsValid() && suggestion.ReturnType != "" {
			// No return type provided: Insert return type before method name
			returnType := methodgeneration.ConcatTypeName(strings.Split(suggestion.ReturnType, " "))
			returnTypeTextEdit := ls.createTextEdit(returnType+" ", lsp.Range{
				Start: doc.ToPosition(method.Name.Range.Start),
				End:   doc.ToPosition(method.Name.Range.Start),
			})
			items[i] = ls.createCompletionItem(parameterListTextEdit, returnTypeTextEdit)
		} else {
			items[i] = ls.createCompletionItem(parameterListTextEdit)
		}
	}
	return items, nil
}

func (ls *languageServer) generateParameterLists(method Method) ([][]predictor.MethodValues, errors.Error) {
	// Generate parameter list
	set, err := ls.findDataset(configuration.LanguageServerMethodGenerationDataset(), predictor.MethodGenerator)
	if err != nil {
		return nil, err
	}

	return predictor.OnDataset(set).GenerateMethods([]predictor.MethodContext{{
		MethodName: method.Name.Content,
		ClassName:  []string{method.ClassName},
		IsStatic:   method.IsStatic,
	}})
}

func (ls *languageServer) createCompletionItem(textEdits ...lsp.TextEdit) lsp.CompletionItem {
	if len(textEdits) == 0 {
		panic("No text edits specified")
	}
	label := textEdits[0].NewText
	if label == "" {
		label = "(no parameters)"
	}

	item := lsp.CompletionItem{
		Label:            label,
		Kind:             lsp.Text,
		Preselect:        true,
		InsertTextFormat: lsp.ITF_PlainText,
		InsertTextMode:   lsp.AsIs,
		SortText:         label,
		FilterText:       label,
	}
	item.TextEdit = &textEdits[0]
	if len(textEdits) >= 1 {
		item.AdditionalTextEdits = textEdits[1:]
	}

	return item
}

func (ls *languageServer) createTextEdit(text string, r lsp.Range) lsp.TextEdit {
	return lsp.TextEdit{
		NewText: text,
		Range:   r,
	}
}

func (ls *languageServer) joinParameterList(value []predictor.Parameter) string {
	output := ""
	for i, par := range value {
		if i > 0 {
			output += ", "
		}
		if par.IsArray {
			output += fmt.Sprintf("%s[] %s", par.Type, par.Name)
		} else {
			output += fmt.Sprintf("%s %s", par.Type, par.Name)
		}
	}
	return output
}

func (ls *languageServer) findDataset(datasetReference string, modelType predictor.SupportedModels) (configuration.Dataset, errors.Error) {
	set, err := configuration.FindDatasetByReference(datasetReference)
	if err != nil {
		models, err := predictor.Global().GetModels(modelType)
		if err != nil {
			return set, err
		}
		for _, model := range models {
			if model.ModelName == datasetReference {
				return configuration.Dataset{
					DatasetBase: configuration.DatasetBase{
						NameRaw: datasetReference,
						PreprocessingOptions: configuration.PreprocessingOptions{
							SentenceFormatting: configuration.SentenceFormattingOptions(model.SentenceFormattingOptions),
						},
						ModelOptions: configuration.ModelOptions{
							ModelType: model.ModelType,
						},
					},
				}, nil
			}
		}
		return set, err
	}
	return set, nil
}

// Logs a message to the default log output.
// @ServiceGenerator:IgnoreMethod
func (ls *languageServer) log(format string, args ...interface{}) {
	log.Print(log.LanguageServer, fmt.Sprintf("[LANGUAGE SERVER] %s\n", format), args...)
}

func (ls *languageServer) IsReturntypeValidationActive() bool {
	//return configuration.LanguageServerReturntypesDataset() != ""
	// No time to test returntype validation, therefore deactivated
	return false
}

func (ls *languageServer) IsMethodGenerationActive() bool {
	return configuration.LanguageServerMethodGenerationDataset() != ""
}

// Creates a progress with token
func (ls *languageServer) CreateProgress(token string) errors.Error {
	if !ls.configuration.IsProgressCreationSupported() {
		return errors.New("Error", "Client does not support progress creation")
	}
	return remote().CreateProgress(token)
}
