/*
   Copyright (c) 2023, btnmasher
   All rights reserved.
   Use of this source code is governed by a BSD-style
   license that can be found in the LICENSE file.
*/

package logfmt

type StyleOption interface {
	apply(*StyleConfig)
}

type scopt func(*StyleConfig)

func (o scopt) apply(f *StyleConfig) {
	o(f)
}

type StyleConfig struct {
	PanicForeground Color
	PanicBackground Color
	PanicStyle      TextStyle
	FatalForeground Color
	FatalBackground Color
	FatalStyle      TextStyle
	ErrorForeground Color
	ErrorBackground Color
	ErrorStyle      TextStyle
	WarnForeground  Color
	WarnBackground  Color
	WarnStyle       TextStyle
	InfoForeground  Color
	InfoBackground  Color
	InfoStyle       TextStyle
	DebugForeground Color
	DebugBackground Color
	DebugStyle      TextStyle
	TraceForeground Color
	TraceBackground Color
	TraceStyle      TextStyle
}

var defaultStyle = StyleConfig{
	PanicForeground: ANSIBrightWhite,
	PanicBackground: ANSIBrightRed,
	PanicStyle:      TextStyle{}.Bold().Blink(),
	FatalForeground: ANSIBrightRed,
	FatalStyle:      TextStyle{}.Bold(),
	ErrorForeground: ANSIRed,
	ErrorStyle:      TextStyle{}.Bold(),
	WarnForeground:  ANSIYellow,
	WarnStyle:       TextStyle{}.Bold(),
	InfoForeground:  ANSICyan,
	InfoStyle:       TextStyle{}.Bold(),
	DebugForeground: ANSIGreen,
	DebugStyle:      TextStyle{}.Bold(),
	TraceForeground: ANSIWhite,
	TraceStyle:      TextStyle{}.Bold(),
}

func NewStyle(options ...StyleOption) StyleConfig {
	style := defaultStyle

	for _, opt := range options {
		opt.apply(&style)
	}

	return style
}

func WithPanicForeground(color Color) StyleOption {
	return scopt(func(s *StyleConfig) {
		s.PanicForeground = color
	})
}

func WithPanicBackground(color Color) StyleOption {
	return scopt(func(s *StyleConfig) {
		s.PanicBackground = color
	})
}

func WithPanicStyle(style TextStyle) StyleOption {
	return scopt(func(s *StyleConfig) {
		s.PanicStyle = style
	})
}

func WithFatalForeground(color Color) StyleOption {
	return scopt(func(s *StyleConfig) {
		s.FatalForeground = color
	})
}

func WithFatalBackground(color Color) StyleOption {
	return scopt(func(s *StyleConfig) {
		s.FatalBackground = color
	})
}

func WithFatalStyle(style TextStyle) StyleOption {
	return scopt(func(s *StyleConfig) {
		s.FatalStyle = style
	})
}

func WithErrorForeground(color Color) StyleOption {
	return scopt(func(s *StyleConfig) {
		s.ErrorForeground = color
	})
}

func WithErrorBackground(color Color) StyleOption {
	return scopt(func(s *StyleConfig) {
		s.ErrorBackground = color
	})
}

func WithErrorStyle(style TextStyle) StyleOption {
	return scopt(func(s *StyleConfig) {
		s.ErrorStyle = style
	})
}

func WithWarnForeground(color Color) StyleOption {
	return scopt(func(s *StyleConfig) {
		s.WarnForeground = color
	})
}

func WithWarnBackground(color Color) StyleOption {
	return scopt(func(s *StyleConfig) {
		s.WarnBackground = color
	})
}

func WithWarnStyle(style TextStyle) StyleOption {
	return scopt(func(s *StyleConfig) {
		s.WarnStyle = style
	})
}

func WithInfoForeground(color Color) StyleOption {
	return scopt(func(s *StyleConfig) {
		s.InfoForeground = color
	})
}

func WithInfoBackground(color Color) StyleOption {
	return scopt(func(s *StyleConfig) {
		s.InfoBackground = color
	})
}

func WithInfoStyle(style TextStyle) StyleOption {
	return scopt(func(s *StyleConfig) {
		s.InfoStyle = style
	})
}

func WithDebugForeground(color Color) StyleOption {
	return scopt(func(s *StyleConfig) {
		s.DebugForeground = color
	})
}

func WithDebugBackground(color Color) StyleOption {
	return scopt(func(s *StyleConfig) {
		s.DebugBackground = color
	})
}

func WithDebugStyle(style TextStyle) StyleOption {
	return scopt(func(s *StyleConfig) {
		s.DebugStyle = style
	})
}

func WithTraceForeground(color Color) StyleOption {
	return scopt(func(s *StyleConfig) {
		s.TraceForeground = color
	})
}

func WithTraceBackground(color Color) StyleOption {
	return scopt(func(s *StyleConfig) {
		s.TraceBackground = color
	})
}

func WithTraceStyle(style TextStyle) StyleOption {
	return scopt(func(s *StyleConfig) {
		s.TraceStyle = style
	})
}
