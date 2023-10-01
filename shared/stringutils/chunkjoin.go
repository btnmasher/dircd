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

package stringutils

import "bytes"

// ChunkJoinStrings takes a list of individual parameters and joins them to strings
// separated by sep, limited by the maxlength. For each item, if appending the item
// would breach the maxlength, it instead starts to build a new string. Once all
// the strings are built, it returns the list of strings.
func ChunkJoinStrings(maxlength int, sep string, params ...string) []string {
	var buffer bytes.Buffer
	currentLength := 0
	var joined []string
	nextBuffer := false

	for i := range params {
		// Check if we have enough room to write the item
		if currentLength+len(params[i]) <= maxlength {
			buffer.WriteString(params[i])
			currentLength += len(params[i])
		} else { // Not enough room, reiterate for the next item
			nextBuffer = true
		}

		// Check if we're currently on the last item or if we can add another with a separator
		if i+1 < len(params) && currentLength+len(sep)+len(params[i+1]) <= maxlength {
			buffer.WriteString(sep)
			currentLength += len(sep)
		} else { // Last item, or we can't fit the next item with a separator
			nextBuffer = true
		}

		if nextBuffer {
			currentLength = 0
			nextBuffer = false
			joined = append(joined, buffer.String())
			buffer.Reset()
		}
	}

	if buffer.Len() > 0 { // Finished iterating without hitting max length on the current pass.
		joined = append(joined, buffer.String())
	}

	return joined
}
