/*
   Copyright (c) 2023, btnmasher
   All rights reserved.
   Use of this source code is governed by a BSD-style
   license that can be found in the LICENSE file.
*/

package itempool

type ScrubbableItem interface {
	Scrub()
}

type InitItemFunc[T ScrubbableItem] func() T

type Pool[T ScrubbableItem] interface {
	Warmup(num int)
	New() T
	Recycle(T)
}

func New[T ScrubbableItem](max int, init InitItemFunc[T]) Pool[T] {
	return &poolImpl[T]{
		queue:    make(chan T, max),
		itemInit: init,
	}
}

type poolImpl[T ScrubbableItem] struct {
	queue    chan T
	itemInit InitItemFunc[T]
}

func (p *poolImpl[T]) Warmup(num int) {
	for i := 0; i < num; i++ {
		select {
		case p.queue <- p.itemInit():
			// nop
		default:
			return
		}
	}
}

func (p *poolImpl[T]) New() (item T) {
	select {
	case item = <-p.queue:
	default:
		item = p.itemInit()
	}
	return
}

func (p *poolImpl[T]) Recycle(item T) {
	item.Scrub()
	select {
	case p.queue <- item:
	default:
		// let it go, let it go...
	}
}
