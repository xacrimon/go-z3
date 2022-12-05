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

func (l String) Replace(what, with String) Value {
	ctx := l.ctx
	val := wrapValue(ctx, func() C.Z3_ast {
		return C.Z3_mk_seq_replace(ctx.c, l.c, what.c, with.c)
	})
	runtime.KeepAlive(l)
	runtime.KeepAlive(what)
	runtime.KeepAlive(with)
	return String(val)
}

func (l String) Length() Int {
	ctx := l.ctx
	val := wrapValue(ctx, func() C.Z3_ast {
		return C.Z3_mk_seq_length(ctx.c, l.c)
	})
	runtime.KeepAlive(l)
	return Int(val)
}

func (l String) Concat(other ...String) String {
	args := []C.Z3_ast{l.c}
	for _, o := range other {
		args = append(args, o.c)
	}

	ctx := l.ctx
	val := wrapValue(ctx, func() C.Z3_ast {
		return C.Z3_mk_seq_concat(ctx.c, C.uint(len(args)), &args[0])
	})
	runtime.KeepAlive(l)
	runtime.KeepAlive(args)
	return String(val)
}

func (l String) Substring(offset, length Value) String {
	ctx := l.ctx
	val := wrapValue(ctx, func() C.Z3_ast {
		return C.Z3_mk_seq_extract(ctx.c, l.c, offset.impl().c, length.impl().c)
	})
	runtime.KeepAlive(l)
	runtime.KeepAlive(offset)
	runtime.KeepAlive(length)
	return String(val)
}

func (l String) ToCode() Int {
	ctx := l.ctx
	val := wrapValue(ctx, func() C.Z3_ast {
		return C.Z3_mk_string_to_code(ctx.c, l.c)
	})
	runtime.KeepAlive(l)
	return Int(val)
}

func (ctx *Context) StringFromCode(c Int) String {
	val := wrapValue(ctx, func() C.Z3_ast {
		return C.Z3_mk_string_from_code(ctx.c, c.c)
	})
	runtime.KeepAlive(c)
	return String(val)
}
