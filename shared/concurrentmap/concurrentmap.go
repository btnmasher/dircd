/*
   Copyright (c) 2023, btnmasher
   All rights reserved.

   Redistribution and use in source and binary forms, with or without modification, are permitted provided that
   the following conditions are met:

   1. Redistributions of source code must retain the above copyright notice, this list of conditions and the
      following disclaimer.

   2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and
      the following disclaimer in the documentation and/or other materials provided with the distribution.

   3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or
      promote products derived from this software without specific prior written permission.

   THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED
   WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A
   PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
   ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED
   TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
   HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
   NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
   POSSIBILITY OF SUCH DAMAGE.
*/

package concurrentmap

import (
	"errors"
	"sync"
)

type ConcurrentMap[K comparable, V any] interface {
	Length() int
	Get(K) (V, bool)
	Set(K, V)
	ChangeKey(K, K) bool
	Delete(K) bool
	Exists(K) bool
	Keys() []K
	Values() []V
	KeysIter() <-chan K
	ValuesItr() <-chan V
	ForEach(func(K, V) error) error
	Clear()
}

func New[K comparable, V any]() ConcurrentMap[K, V] {
	return &concurrentMapImpl[K, V]{
		m: make(map[K]V),
	}
}

type concurrentMapImpl[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

func (cm *concurrentMapImpl[K, V]) Length() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.m)
}

func (cm *concurrentMapImpl[K, V]) Get(key K) (V, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	value, ok := cm.m[key]
	return value, ok
}

func (cm *concurrentMapImpl[K, V]) Set(key K, value V) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.m[key] = value
}

func (cm *concurrentMapImpl[K, V]) ChangeKey(oldKey K, newKey K) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if value, exists := cm.m[oldKey]; exists {
		delete(cm.m, oldKey)
		cm.m[newKey] = value
		return true
	}
	return false
}

func (cm *concurrentMapImpl[K, V]) Delete(key K) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if _, exists := cm.m[key]; exists {
		delete(cm.m, key)
		return true
	}
	return false
}

func (cm *concurrentMapImpl[K, V]) Exists(key K) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	_, ok := cm.m[key]
	return ok
}

func (cm *concurrentMapImpl[K, V]) Keys() []K {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	keys := make([]K, 0, len(cm.m))
	for k := range cm.m {
		keys = append(keys, k)
	}
	return keys
}

func (cm *concurrentMapImpl[K, V]) Values() []V {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	vals := make([]V, 0, len(cm.m))
	for _, v := range cm.m {
		vals = append(vals, v)
	}
	return vals
}

func (cm *concurrentMapImpl[K, V]) KeysIter() <-chan K {
	cm.mu.RLock()
	ch := make(chan K, len(cm.m))
	go func() {
		defer close(ch)
		defer cm.mu.RUnlock()
		for k := range cm.m {
			ch <- k
		}
	}()
	return ch
}

func (cm *concurrentMapImpl[K, V]) ValuesItr() <-chan V {
	cm.mu.RLock()
	ch := make(chan V, len(cm.m))
	go func() {
		defer close(ch)
		defer cm.mu.RUnlock()
		for _, v := range cm.m {
			ch <- v
		}
	}()
	return ch
}

func (cm *concurrentMapImpl[K, V]) ForEach(do func(K, V) error) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	var errs error
	for k, v := range cm.m {
		if err := do(k, v); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return errs
}

func (cm *concurrentMapImpl[K, V]) Clear() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	clear(cm.m)
}
