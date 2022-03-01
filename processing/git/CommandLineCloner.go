package git

import (
	"fmt"
	"io"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/common/utils"
	"time"
)

type Options struct {
	OutputDir string
	URI       string
	Filter    *Filter
}

type Filter struct {
	SizeLimit string
}

func (o Options) toArgs() []string {
	args := make([]string, 0, 4)
	args = append(args, "clone")
	if o.Filter != nil {
		args = append(args, o.Filter.toArgument())
	}
	args = append(args, "--progress", "--verbose")
	args = append(args, o.URI, o.OutputDir)
	return args
}

func (filter Filter) toArgument() string {
	if len(filter.SizeLimit) > 0 {
		return fmt.Sprintf("--filter=blob:limit=%s", filter.SizeLimit)
	}
	return ""
}

// Uses the git command line tool for cloning repositories if it is available on the system.
// This allows use of features that are not available in the current go-git version. (like --filter)
// Also, it seems to be faster.
type CommandLineCloner struct{}

func (c *CommandLineCloner) Clone(uri, outputDir string) errors.Error {
	if err := c.clone(Options{
		URI:       uri,
		OutputDir: outputDir,
		Filter: &Filter{
			SizeLimit: "256k",
		},
	}); err != nil {
		return err
	}
	return nil
}

func (c *CommandLineCloner) clone(options Options) errors.Error {
	process := utils.NewProcess("git", options.toArgs()...)
	stderr, err := process.Stderr()
	if err != nil {
		return err
	}
	defer stderr.Close()

	if err := process.Start(); err != nil {
		return err
	}
	go c.displayGitProgressReports(stderr)
	return process.Wait()
}

func (c *CommandLineCloner) displayGitProgressReports(reader io.Reader) {
	buffer := make([]byte, 256)
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			fmt.Print(string(buffer[:n]))
		}
		if err != nil {
			fmt.Print("\n")
			log.Error(errors.Wrap(err, "Error", "error"))
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	fmt.Print("\n")
}
