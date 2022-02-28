package external

import (
	"returntypes-langserver/common/debug/errors"
)

var gitInstance *Git

func CloneRepository(url, outputDir string) errors.Error {
	if gitInstance == nil {
		gitInstance = &Git{}
		if err := gitInstance.Clone(Options{
			URI:       url,
			OutputDir: outputDir,
			Filter: &Filter{
				SizeLimit: "256k",
			},
		}); err != nil {
			return err
		}
	}
	return nil
}
