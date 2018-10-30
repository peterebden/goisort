// Package isort implements an import sorter & grouper for Go.
// This currently formats to a single style, with three groups
// (stdlib, third-party and local) separated by newlines.
package isort

import (
	"bufio"
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
	StartLine     int      // Line that imports begin on, 1-indexed.
	EndLine       int      // Line that imports end on
	Imports       []Import // List of imports, in order.
	ChangesNeeded bool     // True if changes are needed to this file.
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
)

// Reformat reformats an existing file and returns the details of changes to be made.
func Reformat(filename, localPkg string) (*Changes, error) {
	fset := token.FileSet{}
	f, err := parser.ParseFile(&fset, filename, nil, parser.ImportsOnly)
	if err != nil {
		return err
	}
	changes := &Changes{}
	for _, spec := range f.ImportSpec {
		if changes.StartLine == 0 {
			changes.StartLine = fset.Position(spec.Pos()).Line
		}
		changes.EndLine = fset.Position(spec.EndPos).Line
		changes.Imports = append(Import{
			Path:    spec.Path.Value,
			Name:    spec.Name.Name,
			Doc:     convertComment(spec.Doc),
			Comment: strings.Join(convertComment(spec.Comment), " "),
		})
	}
	// Keep a copy of the original so we can work out if it's changed later.
	imps := changes.Imports
	original := make([]Import, len(imps))
	copy(original, imps)

	stdPkgs := make(map[string]struct{}, len(stdlib))
	for _, pkg := range stdlib {
		stdPkgs[pkg] = struct{}{}
	}

	cmp := func(a, b int) bool {
		typeA := classifyPkg(imps[a].Path)
		typeB := classifyPkg(imps[b].Path)
		if typeA != typeB {
			return typeA < typeB
		} else if imps[a].Path != imps[b].Path {
			return imps[a].Path < imps[b].Path
		}
		return imps[a].Name < imps[b].Name
	}
	changes.ChangesNeeded = sort.SliceIsSorted(imps, cmp)
	if changes.ChangesNeeded {
		sort.Slice(imps, cmp)
	}
	return changes
}

// Rewrite rewrites the contents of a file based on a set of changes.
func Rewrite(filename string, changes *Changes) error {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	lines := strings.Split(string(b), "\n")
	if len(lines) < changes.EndLine {
		return fmt.Errorf("Mismatching file lengths; expected at least %d but got %d", changes.EndLine, len(lines))
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	for i := 0; i < changes.StartLine-1; i++ {
		w.Write(lines[i])
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
		w.Write(lines[i])
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

// classifyPkg classifies a package into one of three buckets; standard library, third-party and local.
func classifyPkg(name, localPkg string, stdPkgs map[string]struct{}) packageType {
	if _, present := stdPkgs[name]; present {
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
func writeImport(w bufio.Writer, imp Import, prefix string) {
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
	w.WriteRune('"')
	w.WriteString(imp.Path)
	w.WriteRune('"')
	w.WriteRune('\n')
}
