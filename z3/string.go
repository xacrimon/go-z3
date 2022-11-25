// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package z3

/*
#include <z3.h>
#include <stdlib.h>
*/
import "C"
import "runtime"

type String value

func init() {
	kindWrappers[KindString] = func(x value) Value {
		return String(x)
	}
}

func (ctx *Context) StringSort() Sort {
	var sort Sort
	ctx.do(func() {
		sort = wrapSort(ctx, C.Z3_mk_string_sort(ctx.c), KindString)
	})
	return sort
}

func (ctx *Context) StringConst(name string) String {
	return ctx.Const(name, ctx.StringSort()).(String)
}

func (l String) Eq(r String) Bool {
	ctx := l.ctx
	val := wrapValue(ctx, func() C.Z3_ast {
		return C.Z3_mk_eq(ctx.c, l.c, r.c)
	})
	runtime.KeepAlive(l)
	runtime.KeepAlive(r)
	return Bool(val)
}
