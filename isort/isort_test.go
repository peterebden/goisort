package isort

import (
	"io/ioutil"
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

func TestRewrite2(t *testing.T) {
	changes, err := Reformat("isort/test_data/test2.go", "")
	assert.NoError(t, err)
	err = Rewrite("isort/test_data/test2.go", "test2_reformatted.go", changes)
	assert.NoError(t, err)
	assertFilesEqual(t, "isort/test_data/test2_reformatted.go", "test2_reformatted.go")
}

func assertFilesEqual(t *testing.T, filename1, filename2 string) {
	b1, err := ioutil.ReadFile(filename1)
	assert.NoError(t, err)
	b2, err := ioutil.ReadFile(filename2)
	assert.NoError(t, err)
	assert.Equal(t, b1, b2)
}
