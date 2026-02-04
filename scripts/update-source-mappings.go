package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type SourceCodeMapping struct {
	SpanName     string `json:"span_name"`
	FilePath     string `json:"file_path"`
	FunctionName string `json:"function_name"`
	StartLine    int    `json:"start_line"`
	EndLine      int    `json:"end_line"`
	Description  string `json:"description,omitempty"`
}

type MappingFile struct {
	Mappings []SourceCodeMapping `json:"mappings"`
}

func main() {
	root := flag.String("root", ".", "repo root to scan")
	out := flag.String("out", "source_code_mappings.json", "output mappings file (relative to root if not absolute)")
	flag.Parse()

	rootAbs, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve root: %v\n", err)
		os.Exit(1)
	}

	outPath := *out
	if !filepath.IsAbs(outPath) {
		outPath = filepath.Join(rootAbs, outPath)
	}

	existingDescriptions := loadDescriptions(outPath)
	mappings, skipped := scanMappings(rootAbs, existingDescriptions)

	sort.Slice(mappings, func(i, j int) bool {
		if mappings[i].FilePath != mappings[j].FilePath {
			return mappings[i].FilePath < mappings[j].FilePath
		}
		if mappings[i].StartLine != mappings[j].StartLine {
			return mappings[i].StartLine < mappings[j].StartLine
		}
		return mappings[i].SpanName < mappings[j].SpanName
	})

	if err := writeMappings(outPath, mappings); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write mappings: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Updated %d mappings -> %s\n", len(mappings), outPath)
	if len(skipped) > 0 {
		fmt.Fprintf(os.Stderr, "Skipped %d dynamic spans (non-literal names):\n", len(skipped))
		for _, item := range skipped {
			fmt.Fprintf(os.Stderr, "- %s\n", item)
		}
	}
}

func loadDescriptions(path string) map[string]string {
	descriptions := make(map[string]string)
	file, err := os.Open(path)
	if err != nil {
		return descriptions
	}
	defer file.Close()

	var existing MappingFile
	if err := json.NewDecoder(file).Decode(&existing); err != nil {
		return descriptions
	}

	for _, mapping := range existing.Mappings {
		if strings.TrimSpace(mapping.Description) != "" {
			descriptions[mapping.SpanName] = mapping.Description
		}
	}

	return descriptions
}

func scanMappings(rootAbs string, existingDescriptions map[string]string) ([]SourceCodeMapping, []string) {
	var mappings []SourceCodeMapping
	var skipped []string
	seen := make(map[string]SourceCodeMapping)

	fset := token.NewFileSet()

	_ = filepath.WalkDir(rootAbs, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			switch d.Name() {
			case ".git", "vendor", "tempo-data", "docs":
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		fileAst, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return nil
		}

		ast.Inspect(fileAst, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				return true
			}

			startLine := fset.Position(fn.Pos()).Line
			endLine := fset.Position(fn.End()).Line

			ast.Inspect(fn.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}

				if !isTracerStart(call) || len(call.Args) < 2 {
					return true
				}

				spanName, ok := literalString(call.Args[1])
				if !ok {
					skipped = append(skipped, fmt.Sprintf("%s:%d in %s", filepath.Base(path), startLine, fn.Name.Name))
					return true
				}

				relPath, err := filepath.Rel(rootAbs, path)
				if err != nil {
					relPath = path
				}

				description := existingDescriptions[spanName]
				if description == "" {
					description = docSummary(fn.Doc)
				}

				mapping := SourceCodeMapping{
					SpanName:     spanName,
					FilePath:     filepath.ToSlash(relPath),
					FunctionName: fn.Name.Name,
					StartLine:    startLine,
					EndLine:      endLine,
					Description:  description,
				}

				if _, exists := seen[spanName]; !exists {
					seen[spanName] = mapping
					mappings = append(mappings, mapping)
				}

				return true
			})

			return true
		})

		return nil
	})

	return mappings, uniqueStrings(skipped)
}

func isTracerStart(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel == nil || sel.Sel.Name != "Start" {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	return ok && ident.Name == "tracer"
}

func literalString(expr ast.Expr) (string, bool) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	value, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", false
	}
	return value, true
}

func docSummary(doc *ast.CommentGroup) string {
	if doc == nil {
		return ""
	}
	text := strings.TrimSpace(doc.Text())
	if text == "" {
		return ""
	}
	if idx := strings.Index(text, "\n"); idx >= 0 {
		text = text[:idx]
	}
	return text
}

func writeMappings(path string, mappings []SourceCodeMapping) error {
	tmpPath := path + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(MappingFile{Mappings: mappings}); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}

func uniqueStrings(items []string) []string {
	seen := make(map[string]struct{})
	unique := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		unique = append(unique, item)
	}
	sort.Strings(unique)
	return unique
}
