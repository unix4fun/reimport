// +build go1.11

package main

// Copyright 2017 (c) Eric "eau" Augé <eau+reimport [A.T.] unix4fun [DOT] net>
//
// Redistribution and use in source and binary forms, with or without modification,
// are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice, this
// list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
// this list of conditions and the following disclaimer in the documentation and/or
// other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its contributors
// may be used to endorse or promote products derived from this software without
// specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
// ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

/*
      __
     (__)_
     (____)_          <-- this is an hermit crab
     (______)
...//(  00 )\.....

Eric "eau" Augé <eau+reimport (AT} unix4fun [D.O.T) net>
pragmatic tools for pragmatic situations without all the opensource circus marketing/logos/etc....

to match/replace in "import" lines in big projects, i did not find something else,
so i wrote a quick stuff using go parser to identify import lines in go files and create a patch
displayed out.

the tool yells a "patch" file, so you can review the changes before applying (and catch bugs).

example :

// patch an entire directory tree
reimport -m hash -r myhash /path/to/dir > import.patch

// patch a few files
reimport -m hash -r myhash /path/to/file.go /path/to/file2.go /path/to/file3.go > import.patch

*/

import (
	"bufio"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"strings"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: reimport [flags] [path ...]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func fileImportPatch(filepath, match, replace string) error {

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filepath, nil, parser.ImportsOnly)
	if err != nil {
		return err
	}

	linemap := fileImportMatchLines(fset, f, match)
	//fmt.Printf("matched lines: %v\n", linemap)
	// patch
	singleFileImportPatch(filepath, match, replace, linemap)
	return nil
}

func dirImportPatch(dirpath, match, replace string) error {
	// scan the directory to find other directories
	// recursive call myself on those new directories.
	// we can make it entirely concurrent later..
	// could use the filter, but well.. i guess i m lazy..

	direntries, err := ioutil.ReadDir(dirpath)
	if err != nil {
		return err
	}

	for _, e := range direntries {
		if e.IsDir() {
			subdir := fmt.Sprintf("%s/%s", dirpath, e.Name())
			err := dirImportPatch(subdir, match, replace)
			if err != nil {
				fmt.Printf("error dir: %s\n", subdir)
			}
		}
	}

	// then let's act on that directory
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dirpath, nil, parser.ImportsOnly)
	if err != nil {
		return err
	}
	if len(pkgs) <= 0 {
		return nil
	}

	for _, p := range pkgs {
		// let's go to the files.
		for fname, f := range p.Files {
			// match lines
			linemap := fileImportMatchLines(fset, f, match)
			// patch out
			singleFileImportPatch(fname, match, replace, linemap)
		}
	}

	return nil
}

func fileImportMatchLines(fset *token.FileSet, f *ast.File, match string) map[int]bool {
	lines := make(map[int]bool)
	for _, i := range f.Imports {
		fpos := fset.Position(i.Pos())
		//fmt.Printf("import: %s off: %d line: %d\n", i.Path.Value, i.Pos(), fpos.Line)
		if strings.Contains(i.Path.Value, match) {
			lines[fpos.Line] = true
		}
	}
	return lines
}

// naive function
func singleFileImportPatch(filename, match, replace string, lines map[int]bool) {
	if len(lines) <= 0 {
		return
	}

	scanFile, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer scanFile.Close()
	scanner := bufio.NewScanner(scanFile)

	//max := lines[len(lines)]
	//fmt.Printf("# matches %d\n", len(lines))

	fmt.Printf("--- %s\n", filename)
	fmt.Printf("+++ %s\n", filename)
	// line array is sorted by nature as we process sequentially.
	l := 1
	for m := 0; m < len(lines); l++ {
		//fmt.Printf("line: %d\n", l)
		scanner.Scan()
		_, ok := lines[l]
		if !ok {
			continue // next line
		}
		origLine := scanner.Text()
		newLine := strings.Replace(origLine, match, replace, 1)

		fmt.Printf("@@ -%d,1 +%d,1 @@\n", l, l)
		fmt.Printf("-%s\n", origLine)
		fmt.Printf("+%s\n", newLine)
		m++ // we've matched
	}
}

// we will spit out a patch file
// to pipe in patch(1)
func main() {
	// let's match.
	matchFlag := flag.String("m", "", "string to match/replace")
	replaceFlag := flag.String("r", "", "string replacement")

	flag.Usage = usage
	flag.Parse()
	argv := flag.Args()

	if len(*matchFlag) == 0 || len(*replaceFlag) == 0 || len(argv) == 0 {
		fmt.Printf("invalid argument\n")
		usage()
		os.Exit(1)
	}

	for _, filepath := range argv {
		fi, err := os.Stat(filepath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid file: %s/%v\n", filepath, err)
			continue
		}

		if fi.IsDir() {
			dirImportPatch(filepath, *matchFlag, *replaceFlag)
		} else {
			fileImportPatch(filepath, *matchFlag, *replaceFlag)
		}
	}

	os.Exit(0)
}
