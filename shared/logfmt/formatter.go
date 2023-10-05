/*
   Copyright (c) 2023, btnmasher
   All rights reserved.
   Use of this source code is governed by a BSD-style
   license that can be found in the LICENSE file.
*/

package logfmt

import (
	"bytes"
	"fmt"
	"io"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/muesli/termenv"
	"github.com/sirupsen/logrus"
)

// Formatter implement logurs.Formatter, offering options to format log entries with nested fields
// with text color and styles
type Formatter struct {
	fieldsOrder           []string
	timestampFormat       string
	hideKeys              bool
	noStyles              bool
	noFieldStyles         bool
	noFieldsSpace         bool
	showFullLevel         bool
	noUppercaseLevel      bool
	trimMessages          bool
	callerFirst           bool
	styleConfig           *StyleConfig
	customCallerFormatter func(*runtime.Frame) string
}

type FormatOption interface {
	apply(*Formatter)
}

type fmtopt func(*Formatter)

func (o fmtopt) apply(f *Formatter) {
	o(f)
}

func New(options ...FormatOption) *Formatter {
	style := defaultStyle
	formatter := &Formatter{
		styleConfig: &style,
	}

	for _, opt := range options {
		opt.apply(formatter)
	}

	return formatter
}

// WithFieldsOrder sets the field display order
// default: fields sorted alphabetically
func WithFieldsOrder(fields ...string) FormatOption {
	return fmtopt(func(f *Formatter) {
		f.fieldsOrder = fields
	})
}

// WithTimestampFormat sets the timestamp format
// default: time.StampMilli = "Jan _2 15:04:05.000"
func WithTimestampFormat(format string) FormatOption {
	return fmtopt(func(f *Formatter) {
		f.timestampFormat = format
	})
}

// HideKeys sets whether to show [fieldValue] instead of [fieldKey:fieldValue]
func HideKeys(state bool) FormatOption {
	return fmtopt(func(f *Formatter) {
		f.hideKeys = state
	})
}

// NoStyles sets whether to disable colors
// default: false
func NoStyles(state bool) FormatOption {
	return fmtopt(func(f *Formatter) {
		f.noStyles = state
	})
}

// NoFieldStyles set whether to apply colors only to the level
// default: level & fields
func NoFieldStyles(state bool) FormatOption {
	return fmtopt(func(f *Formatter) {
		f.noFieldStyles = state
	})
}

// NoFieldsSpace sets whether to disable printing spaces between fields
// default: false
func NoFieldsSpace(state bool) FormatOption {
	return fmtopt(func(f *Formatter) {
		f.noFieldsSpace = state
	})
}

// ShowFullLevel sets whether to show a full level [WARNING] instead of [WARN]
// default: false
func ShowFullLevel(state bool) FormatOption {
	return fmtopt(func(f *Formatter) {
		f.showFullLevel = state
	})
}

// NoUppercaseLevel sets whether to disable printing level values in UPPERCASE
// default: false
func NoUppercaseLevel(state bool) FormatOption {
	return fmtopt(func(f *Formatter) {
		f.noUppercaseLevel = state
	})
}

// TrimMessages sets whether to trim whitespaces on messages
// default: false
func TrimMessages(state bool) FormatOption {
	return fmtopt(func(f *Formatter) {
		f.trimMessages = state
	})
}

// CallerFirst sets whether to print caller info first
// default: false
func CallerFirst(state bool) FormatOption {
	return fmtopt(func(f *Formatter) {
		f.callerFirst = state
	})
}

// WithCustomCallerFormatter sets a custom formatter for caller info
// default: none
func WithCustomCallerFormatter(formatter func(*runtime.Frame) string) FormatOption {
	return fmtopt(func(f *Formatter) {
		f.customCallerFormatter = formatter
	})
}

// WithStyleConfig sets a custom color layout for styling the level and fields
func WithStyleConfig(config StyleConfig) FormatOption {
	return fmtopt(func(f *Formatter) {
		f.styleConfig = &config
	})
}

// Format an log entry
func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	loggerOut := entry.Logger.Out
	profile := termenv.NewOutput(loggerOut).ColorProfile()
	levelStyle := f.getStyleByLevel(entry.Level)

	timestampFormat := f.timestampFormat
	if timestampFormat == "" {
		timestampFormat = time.StampMilli
	}

	// output buffer
	buff := &bytes.Buffer{}

	out := termenv.NewOutput(buff, termenv.WithProfile(profile))

	// write time
	out.WriteString(entry.Time.Format(timestampFormat))

	if f.callerFirst {
		f.writeCaller(out, entry)
	}

	out.WriteString(" ")

	// write fields
	if f.fieldsOrder == nil {
		f.writeFields(out, entry, &levelStyle)
	} else {
		f.writeOrderedFields(out, entry, &levelStyle)
	}

	out.WriteString(" ")

	// write message
	if f.trimMessages {
		out.WriteString(strings.TrimSpace(entry.Message))
	} else {
		out.WriteString(entry.Message)
	}

	if !f.callerFirst {
		f.writeCaller(out, entry)
	}

	out.WriteString("\n")

	return buff.Bytes(), nil
}

func (f *Formatter) writeCaller(out io.Writer, entry *logrus.Entry) {
	if entry.HasCaller() {
		if f.customCallerFormatter != nil {
			fmt.Sprintf(f.customCallerFormatter(entry.Caller))
		} else {
			fmt.Fprintf(
				out,
				" (%s:%d %s)",
				entry.Caller.File,
				entry.Caller.Line,
				entry.Caller.Function,
			)
		}
	}
}

func (f *Formatter) formatLevel(entry *logrus.Entry) string {
	var level string
	if f.noUppercaseLevel {
		level = entry.Level.String()
	} else {
		level = strings.ToUpper(entry.Level.String())
	}

	if !f.showFullLevel {
		level = level[:4]
	}

	return fmt.Sprintf("[%s]", level)
}

func (f *Formatter) formatField(entry *logrus.Entry, field string) string {
	if f.hideKeys {
		return fmt.Sprintf("[%v]", entry.Data[field])
	} else {
		return fmt.Sprintf("[%s:%v]", field, entry.Data[field])
	}
}

func (f *Formatter) writeFields(out io.Writer, entry *logrus.Entry, levelStyle *TextStyle) {
	fields := make([]string, 1, len(entry.Data)+1)
	fields[0] = f.formatLevel(entry)

	if len(entry.Data) != 0 {
		fields := make([]string, 0, len(entry.Data))
		for field := range entry.Data {
			fields = append(fields, field)
		}

		sort.Strings(fields)

		for _, field := range fields {
			fields = append(fields, f.formatField(entry, field))
		}
	}

	f.joinAndWriteStyled(out, levelStyle, fields)
}

func (f *Formatter) writeOrderedFields(out io.Writer, entry *logrus.Entry, levelStyle *TextStyle) {
	length := len(entry.Data)
	foundFieldsMap := map[string]bool{}
	fields := make([]string, 1, length+1)
	fields[0] = f.formatLevel(entry)

	for _, field := range f.fieldsOrder {
		if _, ok := entry.Data[field]; ok {
			foundFieldsMap[field] = true
			length--
			fields = append(fields, f.formatField(entry, field))
		}
	}

	if length > 0 {
		notFoundFields := make([]string, 0, length)
		for field := range entry.Data {
			if foundFieldsMap[field] == false {
				notFoundFields = append(notFoundFields, field)
			}
		}

		sort.Strings(notFoundFields)

		for _, field := range notFoundFields {
			fields = append(fields, f.formatField(entry, field))
		}
	}

	f.joinAndWriteStyled(out, levelStyle, fields)
}

func (f *Formatter) joinAndWriteStyled(out io.Writer, levelStyle *TextStyle, fields []string) {
	join := ""
	if !f.noFieldsSpace {
		join = " "
	}

	if f.noFieldStyles {
		fmt.Fprint(out, strings.Join(fields, join))
	} else {
		joined := strings.Join(fields, join)
		levelStyle.WriteStyled(out, joined)
	}
}

func (f *Formatter) getStyleByLevel(level logrus.Level) TextStyle {
	switch level {
	case logrus.PanicLevel:
		return f.styleConfig.PanicStyle.
			background(f.styleConfig.PanicBackground).
			foreground(f.styleConfig.PanicForeground)
	case logrus.FatalLevel:
		return f.styleConfig.FatalStyle.
			background(f.styleConfig.FatalBackground).
			foreground(f.styleConfig.FatalForeground)
	case logrus.ErrorLevel:
		return f.styleConfig.ErrorStyle.
			background(f.styleConfig.ErrorBackground).
			foreground(f.styleConfig.ErrorForeground)
	case logrus.WarnLevel:
		return f.styleConfig.WarnStyle.
			background(f.styleConfig.WarnBackground).
			foreground(f.styleConfig.WarnForeground)
	case logrus.InfoLevel:
		return f.styleConfig.InfoStyle.
			background(f.styleConfig.InfoBackground).
			foreground(f.styleConfig.InfoForeground)
	case logrus.DebugLevel:
		return f.styleConfig.DebugStyle.
			background(f.styleConfig.DebugBackground).
			foreground(f.styleConfig.DebugForeground)
	case logrus.TraceLevel:
		return f.styleConfig.TraceStyle.
			background(f.styleConfig.TraceBackground).
			foreground(f.styleConfig.TraceForeground)
	default:
		return TextStyle{}
	}
}
