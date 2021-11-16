package workspace

import (
	"returntypes-langserver/common/java"
	"returntypes-langserver/languageserver/diagnostics"
)

// Wraps a java code file together with it's diagnostics.
type FileWrapper struct {
	file        *java.CodeFile
	diagnostics diagnostics.DiagnosticContainer
}

func (wrapper *FileWrapper) Path() string {
	if wrapper.file == nil {
		return ""
	}
	return wrapper.file.FilePath
}

func (wrapper *FileWrapper) Exists() bool {
	return wrapper.file != nil
}

func (wrapper *FileWrapper) File() *java.CodeFile {
	return wrapper.file
}

func (wrapper *FileWrapper) Diagnostics() *diagnostics.DiagnosticContainer {
	return &wrapper.diagnostics
}
