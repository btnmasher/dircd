/*
   Copyright (c) 2020, btnmasher
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

package dircd

import (
	"fmt"
	"sync"
)

// ConnMap is a simple map[string]*Conn wrapped with a concurrent-safe API
type ConnMap struct {
	data map[string]*Conn
	sync.RWMutex
}

// NewConnMap initializes and returns a pointer to a new COnnMap instance.
func NewConnMap() *ConnMap {
	m := &ConnMap{
		data: make(map[string]*Conn),
	}
	return m
}

// ForEach will call the provided function for each entry in the UserMap
func (m *ConnMap) ForEach(do func(*Conn)) {
	m.RLock()
	defer m.RUnlock()

	for _, val := range m.data {
		do(val)
	}
}

// Length returns the length of the underlying map.
func (m *ConnMap) Length() int {
	m.RLock()
	defer m.RUnlock()

	return len(m.data)
}

// Add is used to add a key/value to the map.
// Returns an error if the key already exists.
func (m *ConnMap) Add(key string, value *Conn) error {
	m.Lock()
	defer m.Unlock()

	_, exists := m.data[key]

	if exists {
		return fmt.Errorf("ConnMap: Cannot add map entry, key already exists: %q", key)
	}

	m.data[key] = value
	return nil
}

// Del is used to remove a key/value from the map.
// Returns an error if the key does not exist.
func (m *ConnMap) Del(key string) error {
	m.Lock()
	defer m.Unlock()

	_, exists := m.data[key]

	if !exists {
		return fmt.Errorf("UserMap: Cannot delete map entry, key does not exist: %q", key)
	}

	delete(m.data, key)

	return nil
}

// Get is used to get a key/value from the map.
// Returns an error if the key does not exist.
func (m *ConnMap) Get(key string) (*Conn, error) {
	m.RLock()
	defer m.RUnlock()

	v, exists := m.data[key]

	if !exists {
		return nil, fmt.Errorf("UserMap: Cannot get map value, key does not exist: %q", key)
	}

	return v, nil
}

// Set is used to change an existing key/value in the map.
// Returns an error if the key does not exist.
func (m *ConnMap) Set(key string, value *Conn) error {
	m.Lock()
	defer m.Unlock()

	_, exists := m.data[key]

	if !exists {
		return fmt.Errorf("UserMap: Cannot set map value, key does not exist: %q", key)
	}

	m.data[key] = value

	return nil
}

// Exists is used by external callers to check if a value
// exists in the map and returns a boolean with the result.
func (m *ConnMap) Exists(key string) bool {
	m.RLock()
	defer m.RUnlock()

	_, exists := m.data[key]
	return exists
}