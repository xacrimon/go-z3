// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package z3

/*
#include <z3.h>
*/
import "C"
import "runtime"

type Sequence value

func init() {
	kindWrappers[KindSequence] = func(x value) Value {
		return Sequence(x)
	}
}

func (ctx *Context) SequenceSort(element Sort) Sort {
	sort := wrapSort(ctx, C.Z3_mk_seq_sort(ctx.c, element.c), KindSequence)
	runtime.KeepAlive(element)
	return sort
}

func (ctx *Context) SequenceEmpty(sort Sort) Sequence {
	seq := wrapValue(ctx, func() C.Z3_ast {
		qq := C.Z3_mk_seq_empty(ctx.c, sort.c)
		return qq
	})
	runtime.KeepAlive(sort)
	return Sequence(seq)
}

func (ctx *Context) SequenceUnit(item Value) Sequence {
	seq := wrapValue(ctx, func() C.Z3_ast {
		return C.Z3_mk_seq_unit(ctx.c, item.impl().c)
	})
	runtime.KeepAlive(item)
	return Sequence(seq)
}

func (s Sequence) Contains(item Value) Bool {
	ctx := s.ctx
	sub := ctx.SequenceUnit(item)
	val := wrapValue(ctx, func() C.Z3_ast {
		return C.Z3_mk_seq_contains(ctx.c, s.c, sub.impl().c)
	})
	runtime.KeepAlive(s)
	runtime.KeepAlive(item)
	return Bool(val)
}

func (s Sequence) Concat(others ...Sequence) Sequence {
	args := []C.Z3_ast{s.c}
	for _, other := range others {
		args = append(args, other.c)
	}
	ctx := s.ctx
	val := wrapValue(ctx, func() C.Z3_ast {
		return C.Z3_mk_seq_concat(ctx.c, C.uint(len(args)), &args[0])
	})
	runtime.KeepAlive(s)
	runtime.KeepAlive(others)
	return Sequence(val)
}

func (s Sequence) Length() Int {
	val := wrapValue(s.ctx, func() C.Z3_ast {
		return C.Z3_mk_seq_length(s.ctx.c, s.c)
	})
	runtime.KeepAlive(s)
	return Int(val)
}
