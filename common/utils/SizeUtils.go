package utils

import "fmt"

// Size in bytes
type DataSize int64

func Kilobytes(size int) DataSize {
	return DataSize(size * 1024)
}

// Returns the data size in IEC format
func (s DataSize) ToIEC() string {
	// IEC formatting code from:
	// https://programming.guide/go/formatting-byte-size-to-human-readable-format.html
	const unit = DataSize(1024)
	if s < unit {
		return fmt.Sprintf("%d B", s)
	}
	div, exp := unit, 0
	for n := s / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(s)/float64(div), "KMGTPE"[exp])
}

// Given some proportions (each > 0), fits the given integer value to these proportions
func FitProportions(proportionA, proportionB float64, value int) (valueA, valueB int) {
	proportionSum := proportionA + proportionB
	relativeSizeA := proportionA / proportionSum
	valueA = int(float64(value) * relativeSizeA)
	valueB = value - valueA
	return
}
