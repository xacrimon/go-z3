// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"strings"
)

var flagType = flag.String("t", "", "default arguments and results to `type`")

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -t type file.go [file2.go...]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(2)
	}

	if !strings.HasSuffix(flag.Arg(0), ".go") {
		fmt.Fprintf(os.Stderr, "not a .go file: %s\n", flag.Arg(0))
		os.Exit(1)
	}
	nfilename := flag.Arg(0)[:len(flag.Arg(0))-3] + ".wrap.go"

	// Emit prologue.
	var out bytes.Buffer
	fmt.Fprintf(&out, `// Generated by genwrap.go. DO NOT EDIT

package z3

import "runtime"

/*
#cgo LDFLAGS: -lz3
#include <z3.h>
#include <stdlib.h>
*/
import "C"

`)

	// Emit common methods.
	genCommon(&out)

	for _, filename := range flag.Args() {
		code, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// Process the file one line at a time.
		lines := bytes.Split(code, []byte("\n"))

		// Process lines.
		doc := [][]byte{}
		for i, line := range lines {
			if len(line) >= 2 && line[0] == '/' && line[1] == '/' {
				label := fmt.Sprintf("%s:%d", filename, i+1)
				process(&out, line, doc, label)
				doc = append(doc, line)
			} else {
				doc = nil
			}
		}
	}

	// Produce the output code.
	ncode, err := format.Source(out.Bytes())
	if err != nil {
		fmt.Fprintln(os.Stderr, out.String())
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := ioutil.WriteFile(nfilename, ncode, 0666); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func genCommon(w *bytes.Buffer) {
	fmt.Fprintln(w, "// Eq returns an expression that is true if l and r are equal.")
	dir := parseDirective(strings.Fields("//wrap:expr Eq:Bool Z3_mk_eq l r"))
	genMethod(w, dir, "")

	fmt.Fprintf(w, `// NE returns an expression that is true if l and r are not equal.
func (l %s) NE(r %s) Bool {
	return l.ctx.Distinct(l, r)
}

`, *flagType, *flagType)
}

type directive struct {
	goFn, cFn     string
	goArgs, cArgs []*arg
	isDDD         bool
	resType       string
}

type arg struct {
	name, goTyp, cExpr, cCode string
}

func (a arg) c(varName string) string {
	if a.cCode != "" {
		return a.cCode
	}
	return fmt.Sprintf(a.cExpr, varName)
}

func split(x, def string) (a, b string) {
	if i := strings.Index(x, ":"); i >= 0 {
		return x[:i], x[i+1:]
	}
	return x, def
}

func parseDirective(parts []string) *directive {
	defType := *flagType

	colon, cPos := -1, 2
	for i, p := range parts {
		if p == ":" {
			colon, cPos = i, i+1
			break
		}
	}

	goFn, resType := split(parts[1], defType)
	dir := &directive{
		goFn: goFn, cFn: parts[cPos],
		resType: resType,
	}

	cArgs := parts[cPos+1:]
	goArgs := cArgs
	if colon >= 0 {
		goArgs = parts[2:colon]
	}

	argMap := make(map[string]*arg)
	for _, goArg := range goArgs {
		name, goTyp := split(goArg, defType)
		argMap[name] = &arg{name, goTyp, "", ""}
		dir.goArgs = append(dir.goArgs, argMap[name])
	}
	for _, cArg := range cArgs {
		if cArg[0] == '"' {
			// Literal code.
			cCode := cArg[1 : len(cArg)-1]
			dir.cArgs = append(dir.cArgs, &arg{cCode: cCode})
			continue
		}

		name, cTyp := split(cArg, "")
		arg := argMap[name]
		if arg == nil {
			fmt.Fprintf(os.Stderr, "reference to unknown argument %q", name)
			os.Exit(1)
		}
		if cTyp == "" && arg.goTyp == "Expr" {
			arg.cExpr = "%s.impl().c" // Expr interface
		} else if cTyp == "" {
			arg.cExpr = "%s.c" // expr wrapper
		} else {
			arg.cExpr = "C." + cTyp + "(%s)" // basic type
		}
		dir.cArgs = append(dir.cArgs, arg)
	}

	if strings.HasSuffix(dir.goArgs[len(dir.goArgs)-1].name, "...") {
		dir.isDDD = true
	}

	return dir
}

func process(w *bytes.Buffer, line []byte, doc [][]byte, label string) {
	if !bytes.Contains(line, []byte("//wrap:expr")) {
		return
	}
	parts := strings.Fields(string(line))
	if parts[0] != "//wrap:expr" {
		return
	}

	// Found wrap directive.
	dir := parseDirective(parts)

	// Function documentation.
	if len(doc) > 0 && string(doc[len(doc)-1]) == "//" {
		doc = doc[:len(doc)-1]
	}
	for _, line := range doc {
		fmt.Fprintf(w, "%s\n", line)
	}

	genMethod(w, dir, label)
}

func genMethod(w *bytes.Buffer, dir *directive, label string) {
	// Function declaration.
	fmt.Fprintf(w, "func (%s %s) %s(", dir.goArgs[0].name, dir.goArgs[0].goTyp, dir.goFn)
	for i, a := range dir.goArgs[1:] {
		if i > 0 {
			fmt.Fprintf(w, ", ")
		}
		fmt.Fprintf(w, "%s %s", a.name, a.goTyp)
	}
	fmt.Fprintf(w, ") %s {\n", dir.resType)
	if label != "" {
		fmt.Fprintf(w, " // Generated from %s.\n", label)
	}

	if dir.goArgs[0].goTyp != "*Context" {
		// Context is implied by the receiver.
		fmt.Fprintf(w, " ctx := %s.ctx\n", dir.goArgs[0].name)
	}

	if dir.isDDD {
		// Convert arguments to C array.
		arg := dir.cArgs[len(dir.cArgs)-1]
		ddd := arg.name
		ddd = ddd[:len(ddd)-3]
		fmt.Fprintf(w, " cargs := make([]C.Z3_ast, len(%s)+%d)\n", ddd, len(dir.cArgs)-1)
		for i, arg := range dir.cArgs[:len(dir.cArgs)-1] {
			fmt.Fprintf(w, " cargs[%d] = %s\n", i, arg.c(arg.name))
		}
		fmt.Fprintf(w, " for i, arg := range %s { cargs[i+%d] = %s }\n", ddd, len(dir.cArgs)-1, arg.c("arg"))
	}

	// Construct the AST.
	fmt.Fprintf(w, " var cexpr C.Z3_ast\n")
	fmt.Fprintf(w, " ctx.do(func() {\n")
	fmt.Fprintf(w, "  cexpr = C.%s(ctx.c", dir.cFn)
	if !dir.isDDD {
		for _, a := range dir.cArgs {
			fmt.Fprintf(w, ", %s", a.c(a.name))
		}
	} else {
		fmt.Fprintf(w, ", C.uint(len(cargs)), &cargs[0]")
	}
	fmt.Fprintf(w, ")\n")
	fmt.Fprintf(w, " })\n")

	// Keep arguments alive.
	if !dir.isDDD {
		for _, a := range dir.goArgs {
			if a.goTyp != "int" && a.name != "ctx" {
				fmt.Fprintf(w, " runtime.KeepAlive(%s)\n", a.name)
			}
		}
	} else {
		fmt.Fprintf(w, " runtime.KeepAlive(&cargs[0])\n")
	}

	// Wrap the final C result in a Go result.
	expr := "wrapExpr(ctx, cexpr)"
	if dir.resType == "Expr" {
		// Determine the concrete type dynamically.
		fmt.Fprintf(w, " return %s.lift(KindUnknown)", expr)
	} else {
		fmt.Fprintf(w, " return %s(%s)\n", dir.resType, expr)
	}
	fmt.Fprintf(w, "}\n")
	fmt.Fprintf(w, "\n")
}
