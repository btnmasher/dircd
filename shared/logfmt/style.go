/*
	MIT License

	Copyright (c) 2019 Christian Muehlhaeuser

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE.
*/

package logfmt

import (
	"fmt"
	"io"
	"strings"

	"github.com/muesli/termenv"
)

// Sequence definitions.
const (
	// Escape character
	esc = '\x1b'
	// Bell
	bel = '\a'
	// Control Sequence Introducer
	csi = string(esc) + "["
	// Operating System Command
	osc = string(esc) + "]"
	// String Terminator
	st = string(esc) + `\`

	reset     = "0"
	bold      = "1"
	faint     = "2"
	italic    = "3"
	underline = "4"
	blink     = "5"
	reverse   = "7"
	crossOut  = "9"
	overline  = "53"
)

// TextStyle is a collection of styles to be applied to input
type TextStyle struct {
	styles []string
}

// ApplyStyle renders text with all applied styles.
func (t TextStyle) ApplyStyle(text string) string {
	if len(t.styles) == 0 {
		return text
	}

	seq := strings.Join(t.styles, ";")
	if seq == "" {
		return text
	}

	return fmt.Sprintf("%s%sm%s%sm", csi, seq, text, csi+reset)
}

// WriteStyled renders text with all applied styles to the given io.Writer
func (t TextStyle) WriteStyled(out io.Writer, text string) (int, error) {
	if len(t.styles) == 0 {
		return fmt.Fprint(out, text)
	}

	seq := strings.Join(t.styles, ";")
	if seq == "" {
		return fmt.Fprint(out, text)
	}

	return fmt.Fprintf(out, "%s%sm%s%sm", csi, seq, text, csi+reset)
}

func (t TextStyle) WriteStyledf(out io.Writer, format string, a ...any) (int, error) {
	str := fmt.Sprintf(format, a...)
	if len(t.styles) == 0 {
		return fmt.Fprint(out, str)
	}

	seq := strings.Join(t.styles, ";")
	if seq == "" {
		return fmt.Fprint(out, str)
	}

	return fmt.Fprintf(out, "%s%sm%s%sm", csi, seq, str, csi+reset)
}

// foreground sets a foreground color.
func (t TextStyle) foreground(c termenv.Color) TextStyle {
	if c != nil {
		t.styles = append(t.styles, c.Sequence(false))
	}
	return t
}

// background sets a background color.
func (t TextStyle) background(c termenv.Color) TextStyle {
	if c != nil {
		t.styles = append(t.styles, c.Sequence(true))
	}
	return t
}

// Bold enables bold rendering.
func (t TextStyle) Bold() TextStyle {
	t.styles = append(t.styles, bold)
	return t
}

// Faint enables faint rendering.
func (t TextStyle) Faint() TextStyle {
	t.styles = append(t.styles, faint)
	return t
}

// Italic enables italic rendering.
func (t TextStyle) Italic() TextStyle {
	t.styles = append(t.styles, italic)
	return t
}

// Underline enables underline rendering.
func (t TextStyle) Underline() TextStyle {
	t.styles = append(t.styles, underline)
	return t
}

// Overline enables overline rendering.
func (t TextStyle) Overline() TextStyle {
	t.styles = append(t.styles, overline)
	return t
}

// Blink enables blink mode.
func (t TextStyle) Blink() TextStyle {
	t.styles = append(t.styles, blink)
	return t
}

// Reverse enables reverse color mode.
func (t TextStyle) Reverse() TextStyle {
	t.styles = append(t.styles, reverse)
	return t
}

// CrossOut enables crossed-out rendering.
func (t TextStyle) CrossOut() TextStyle {
	t.styles = append(t.styles, crossOut)
	return t
}
