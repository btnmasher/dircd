/*
   Copyright (c) 2023, btnmasher
   All rights reserved.
   Use of this source code is governed by a BSD-style
   license that can be found in the LICENSE file.
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
