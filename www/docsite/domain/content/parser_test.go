package content

import (
	_ "embed"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	//go:embed testdata/codeblocks.md
	testCodeBlocks []byte
)

func Test_parseDocument(t *testing.T) {
	byt := []byte(testCodeBlocks)
	doc, err := parseDocument(byt)
	assert.NoError(t, err)
	assert.NotEmpty(t, doc)
}
