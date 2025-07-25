package lsp

import (
	"returntypes-langserver/common/code/java"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/utils"
	"strings"
)

// Converts the javaparser Range object into a LSP Range object.
func FromJavaRange(javaRange java.Range) Range {
	return Range{
		Start: Position{
			Line:      javaRange.Begin.Line - 1,
			Character: javaRange.Begin.Col - 1,
		},
		End: Position{
			Line:      javaRange.End.Line - 1,
			Character: javaRange.End.Col + 1,
		},
	}
}

// Converts the lsp Range object into a javaparser Range object.
func ToJavaRange(lspRange Range) java.Range {
	return java.Range{
		Begin: java.Position{
			Line: lspRange.Start.Line + 1,
			Col:  lspRange.Start.Character + 1,
		},
		End: java.Position{
			Line: lspRange.End.Line + 1,
			Col:  lspRange.End.Character - 1,
		},
	}
}

// Returns the filepath as a DocumentURI in the file scheme.
func FilePathToDocumentURI(path string) DocumentURI {
	return DocumentURI(utils.FilePathToURI(strings.Replace(path, ":", "%3A", 1)))
}

// Parses a DocumentURI as a local filepath if possible, otherwise returns an error.
func DocumentURIToFilePath(uri DocumentURI) (string, errors.Error) {
	return utils.URIToFilePath(string(uri))
}

// Creates a file operation filter.
func CreateFileOperationFilter(scheme string, matches FileOperationPatternKind, pattern string) FileOperationFilter {
	filter := FileOperationFilter{
		Scheme:  scheme,
		Pattern: CreateFileOperationPattern(pattern, matches),
	}
	return filter
}

// Creates a file operation pattern.
func CreateFileOperationPattern(pattern string, matches FileOperationPatternKind) FileOperationPattern {
	return FileOperationPattern{
		Glob:    pattern,
		Matches: matches,
		Options: &FileOperationPatternOptions{
			IgnoreCase: true,
		},
	}
}

// Returns true if this position comes before the other position.
func (thisPos *Position) IsBefore(otherPos Position) bool {
	return comparePositions(*thisPos, otherPos) == Before
}

// Returns true if this position comes after the other position.
func (thisPos *Position) IsAfter(otherPos Position) bool {
	return comparePositions(*thisPos, otherPos) == After
}

// Returns true if this position and the other position points to the same position.
func (thisPos *Position) IsSame(otherPos Position) bool {
	return comparePositions(*thisPos, otherPos) == Same
}

type ComparisonResult int

const (
	Before ComparisonResult = iota
	After
	Same
)

// Compares pos1 with pos2. The result is describing the relative position of pos1 (so in case its before, it means pos1 comes before pos2).
func comparePositions(pos1, pos2 Position) ComparisonResult {
	if pos1.Line < pos2.Line {
		return Before
	} else if pos1.Line > pos2.Line {
		return After
	} else {
		if pos1.Character < pos2.Character {
			return Before
		} else if pos1.Character > pos2.Character {
			return After
		} else {
			return Same
		}
	}
}

// Returns true if the range is empty (so it is a range with range.start = range.end, length 0)
func (r *Range) IsEmpty() bool {
	return r.Start.IsSame(r.End)
}
