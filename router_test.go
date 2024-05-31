package torque

import (
	"embed"
	"testing"
)

//go:embed README.md
var fsys embed.FS

func Test_Fs(t *testing.T) {
	logFileSystem(fsys)
}
