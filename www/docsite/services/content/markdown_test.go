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

// Test_compile simply tests that all markdown content can be
// parsed and compiled without errors.
func Test_compile(t *testing.T) {

}

func Test_parseDocument(t *testing.T) {
	byt := []byte(testCodeBlocks)
	doc, err := compileMarkdownFile(byt)
	assert.NoError(t, err)
	assert.NotEmpty(t, doc)
}
