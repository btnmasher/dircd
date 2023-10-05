/*
   Copyright (c) 2023, btnmasher
   All rights reserved.
   Use of this source code is governed by a BSD-style
   license that can be found in the LICENSE file.
*/

package pool

import (
	"sync"
)

// Resettable is an interface which defines an item which has a Reset() method defined/
// used for clearing the state of the item before it is recycled to the pool.
type Resettable interface {
	Reset()
}

// A Pool is a generic wrapper around a sync.Pool.
type Pool[T Resettable] struct {
	pool sync.Pool
}

// New creates a new Pool with the provided factory function.
//
// The equivalent sync.Pool construct is "sync.Pool{New: fn}"
func New[T Resettable](factory func() T) Pool[T] {
	return Pool[T]{
		pool: sync.Pool{New: func() any { return factory() }},
	}
}

// New is a generic wrapper around sync.Pool's Get method.
func (p *Pool[T]) New() T {
	return p.pool.Get().(T)
}

// Recycle is a generic wrapper around sync.Pool's Put method, but it first calls .Reset() on the item
func (p *Pool[T]) Recycle(item T) {
	item.Reset()
	p.pool.Put(item)
}
