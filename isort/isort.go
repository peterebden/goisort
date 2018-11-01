// Package isort implements an import sorter & grouper for Go.
// This currently formats to a single style, with three groups
// (stdlib, third-party and local) separated by newlines.
package isort

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

// Changes describes the set of changes requested to a file.
type Changes struct {
	StartLine int      // Line that imports begin on, 1-indexed.
	EndLine   int      // Line that imports end on
	Imports   []Import // List of imports, in order.
	Needed    bool     // True if changes are needed to this file.
}

// An Import describes a single import path.
type Import struct {
	Name    string   // Local name, empty if not set.
	Path    string   // The import path
	Doc     []string // Any preceding comment
	Comment string   // Comment immediately after the import path.
}

type packageType int

const (
	standardLibrary packageType = 0
	thirdParty                  = 1
	localPackage                = 2
	blankLine                   = 3
)

// Reformat reformats an existing file and returns the details of changes to be made.
func Reformat(filename, localPkg string) (*Changes, error) {
	fset := token.FileSet{}
	f, err := parser.ParseFile(&fset, filename, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	changes := &Changes{}
	for i, spec := range f.Imports {
		line := fset.Position(spec.Pos()).Line
		if changes.StartLine == 0 {
			changes.StartLine = line
		}
		if line > changes.EndLine+1 && i > 0 {
			changes.Imports = append(changes.Imports, Import{}) // blank line
		}
		if spec.EndPos == 0 { // Not guaranteed to be set
			spec.EndPos = spec.Path.Pos()
		}
		changes.EndLine = fset.Position(spec.EndPos).Line
		name := ""
		if spec.Name != nil {
			name = spec.Name.Name
		}
		changes.Imports = append(changes.Imports, Import{
			Path:    spec.Path.Value,
			Name:    name,
			Doc:     convertComment(spec.Doc),
			Comment: strings.Join(convertComment(spec.Comment), " "),
		})
	}
	// Keep a copy of the original so we can work out if it's changed later.
	imps := changes.Imports
	original := make([]Import, len(imps))
	copy(original, imps)

	stdPkgs := stdPkgMap()
	cmp := func(a, b int) bool {
		pathA := strings.Trim(imps[a].Path, `"`)
		pathB := strings.Trim(imps[b].Path, `"`)
		typeA := classifyPkg(pathA, localPkg, stdPkgs)
		typeB := classifyPkg(pathB, localPkg, stdPkgs)
		if typeA != typeB {
			return typeA < typeB
		} else if pathA != pathB {
			return pathA < pathB
		}
		return imps[a].Name < imps[b].Name
	}
	sort.Slice(imps, cmp)
	// Add spaces if required
	imps2 := make([]Import, 0, len(imps)+2)
	lastType := standardLibrary
	for i, imp := range imps {
		thisType := classifyPkg(strings.Trim(imp.Path, `"`), localPkg, stdPkgs)
		if thisType != blankLine {
			if thisType != lastType && i != 0 {
				imps2 = append(imps2, Import{})
			}
			imps2 = append(imps2, imp)
		}
		lastType = thisType
	}
	if len(imps2) != len(original) {
		changes.Needed = true
	} else {
		for i, imp := range original {
			if imp.Path != imps2[i].Path || imp.Name != imps2[i].Name {
				changes.Needed = true
				break
			}
		}
	}
	changes.Imports = imps2
	return changes, nil
}

// Rewrite rewrites the contents of a file based on a set of changes.
func Rewrite(infile, outfile string, changes *Changes) error {
	if !changes.Needed {
		return nil
	}
	b, err := ioutil.ReadFile(infile)
	if err != nil {
		return err
	}
	lines := strings.Split(string(b), "\n")
	if len(lines) < changes.EndLine {
		return fmt.Errorf("Mismatching file lengths; expected at least %d but got %d", changes.EndLine, len(lines))
	}
	f, err := os.Create(outfile)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	for i := 0; i < changes.StartLine-1; i++ {
		if strings.HasPrefix(lines[i], "import") {
			break
		}
		w.WriteString(lines[i])
		w.WriteRune('\n')
	}
	if len(changes.Imports) == 1 {
		// Special case to write on a single line.
		writeImport(w, changes.Imports[0], "")
	} else {
		w.WriteString("import (\n")
		for _, imp := range changes.Imports {
			writeImport(w, imp, "\t")
		}
		w.WriteRune('\n')
	}
	for i := changes.EndLine - 1; i < len(lines); i++ {
		w.WriteString(lines[i])
		w.WriteRune('\n')
	}
	return nil
}

func convertComment(cg *ast.CommentGroup) []string {
	if cg == nil {
		return nil
	}
	ret := make([]string, len(cg.List))
	for i, c := range cg.List {
		ret[i] = c.Text
	}
	return ret
}

func stdPkgMap() map[string]struct{} {
	m := make(map[string]struct{}, len(stdlib))
	for _, pkg := range stdlib {
		m[pkg] = struct{}{}
	}
	return m
}

// classifyPkg classifies a package into one of three buckets; standard library, third-party and local.
func classifyPkg(name, localPkg string, stdPkgs map[string]struct{}) packageType {
	if name == "" {
		return blankLine
	} else if _, present := stdPkgs[name]; present {
		return standardLibrary
	} else if localPkg != "" && strings.HasPrefix(name, localPkg) {
		return localPackage
	} else if strings.ContainsRune(name, '.') {
		// TODO(peter): this is a little dodgy as a derivation of what counts as
		//              "third-party", but in practice the dot is a pretty good identifier.
		return thirdParty
	}
	// It's not standard library or obviously third-party, assume it must be local.
	return localPackage
}

// writeImport writes a single import to the given writer.
func writeImport(w *bufio.Writer, imp Import, prefix string) {
	for _, doc := range imp.Doc {
		w.WriteString(prefix)
		w.WriteString(doc)
		w.WriteRune('\n')
	}
	w.WriteString(prefix)
	if imp.Name != "" {
		w.WriteString(imp.Name)
		w.WriteRune(' ')
	}
	w.WriteString(imp.Path)
	w.WriteRune('\n')
}
