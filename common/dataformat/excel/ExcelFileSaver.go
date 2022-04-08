package excel

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"regexp"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/utils"
	"strings"

	"github.com/xuri/excelize/v2"
)

var ErrCouldNotSave = errors.ErrorId("Excel", "Could not save excel file")

// Saves an excel file and applying a fix to the fonts of chart labels, which appear in light gray color in the current excelize version
// Related issue: https://github.com/qax-os/excelize/issues/320
func SaveFile(file *excelize.File) errors.Error {
	buf := &byteBuffer{}
	if err := file.Write(buf); err != nil {
		return ErrCouldNotSave.New()
	}
	output := utils.OpenFileLazy(file.Path)
	defer output.Close()
	if err := applyChangesToOutputFile(buf, output); err != nil {
		return ErrCouldNotSave.New()
	}
	return nil
}

var chartFileMatcher = regexp.MustCompile("/charts/chart\\d+\\.xml")

func applyChangesToOutputFile(in *byteBuffer, out io.Writer) error {
	w := zip.NewWriter(out)
	reader, err := zip.NewReader(in, int64(in.Len()))
	if err != nil {
		return err
	}
	for _, file := range reader.File {
		if chartFileMatcher.Match([]byte(file.Name)) {
			// Replace any occurences of "lumOff" where it is set to 85000 by setting it to 0 inside the chart files
			// This will make the label colors black.
			if content, err := readZipFileContent(file); err != nil {
				return err
			} else if err := writeFileToZip(w, file.Name, fixLumOff(content)); err != nil {
				return err
			}
		} else {
			// Copy any files to the actual output file, which are not the chart files
			if err := w.Copy(file); err != nil {
				return err
			}
		}
	}
	return w.Close()
}

func readZipFileContent(file *zip.File) ([]byte, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(reader)
}

func writeFileToZip(zipWriter *zip.Writer, name string, content []byte) error {
	fileWriter, err := zipWriter.Create(name)
	if err != nil {
		return err
	}
	_, err = fileWriter.Write(content)
	return err
}

func fixLumOff(content []byte) []byte {
	return []byte(strings.ReplaceAll(string(content), `lumOff val="85000"`, `lumOff val="0"`))
}

// A buffer for bytes which implements the io.Writer and io.ReaderAt interfaces
type byteBuffer struct {
	buf []byte
}

func (b *byteBuffer) Write(p []byte) (n int, err error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

func (b *byteBuffer) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= int64(len(b.buf)) {
		return 0, io.EOF
	}
	return b.readAt(p, int(off))
}

func (b *byteBuffer) readAt(p []byte, off int) (n int, err error) {
	n = copy(p, b.buf[off:])
	if n == len(b.buf[off:]) {
		err = io.EOF
	}
	return
}

func (b *byteBuffer) Len() int {
	return len(b.buf)
}
