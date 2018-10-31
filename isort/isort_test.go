package isort

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReformat1(t *testing.T) {
	changes, err := Reformat("isort/test_data/test1.go", "")
	assert.NoError(t, err)
	assert.False(t, changes.Needed)
}

func TestReformat2(t *testing.T) {
	changes, err := Reformat("isort/test_data/test2.go", "")
	assert.NoError(t, err)
	assert.True(t, changes.Needed)
}

func TestClassifyPkg(t *testing.T) {
	stdPkgs := stdPkgMap()
	assert.Equal(t, standardLibrary, classifyPkg("strings", "", stdPkgs))
}
