// Package cmdbuilder provides an options-to-command-line arguments converter.
//
// It's used to convert different flag sets and struct tag based flag options back to a command-line.
//
// For now, it supports the following options:
//
//	native Go FlagSet
//	https://github.com/spf13/pflag FlagSet
//	https://github.com/jessevdk/go-flags like struct tag mapped struct
package cmdbuilder

import (
	"strconv"
	"strings"
)

// Build converts provided args to command-line arguments using default options
func Build(args []Arg) ([]string, error) {
	return (&Builder{}).Build(args)
}

// Builder represents an options-to-command-line arguments converter and its options
type Builder struct {
	DisableShortOption           bool   // Whether to ignore argument shorthands if its long name is specified
	DisableCombiningShortOptions bool   // Whether to not combine shorthand arguments to one command-line argument
	ShortOptionDelimiter         string // Delimiter is written before argument shorthand
	LongOptionDelimiter          string // Delimiter is written before long argument name
	NameValueDelimiter           string // Delimiter is written between argument name and value
	QuoteValue                   bool   // Whether to quote long string values
}

// Build converts provided args to command-line arguments
func (b *Builder) Build(args []Arg) ([]string, error) {
	var ret []string
	var args2 []Arg
	if !b.DisableCombiningShortOptions && !b.DisableShortOption {
		var shorts []string
		for _, arg := range args {
			if arg.IsOption() && arg.IsProvided() && arg.IsValueOptional() && !arg.IsValueProvided() && arg.ShortName() != "" {
				shorts = append(shorts, arg.ShortName())
			} else {
				args2 = append(args2, arg)
			}
		}
		if len(shorts) > 0 {
			ret = append(ret, stringOrDefault(b.ShortOptionDelimiter, shortOptionDelimiter)+strings.Join(shorts, ""))
		}
	} else {
		args2 = args
	}
	for _, arg := range args2 {
		if !arg.IsOption() || !arg.IsProvided() {
			continue
		}
		var name string
		if b.DisableShortOption || arg.ShortName() == "" {
			name = stringOrDefault(b.LongOptionDelimiter, longOptionDelimiter) + arg.Name()
		} else {
			name = stringOrDefault(b.ShortOptionDelimiter, shortOptionDelimiter) + arg.ShortName()
		}
		if arg.IsValueOptional() && arg.IsValueProvided() {
			name += stringOrDefault(b.NameValueDelimiter, nameValueDelimiter)
		}
		for _, v := range arg.Value() {
			if arg.IsValueOptional() {
				if arg.IsValueProvided() {
					ret = append(ret, name+b.quote(v))
				} else {
					ret = append(ret, name)
				}
			} else {
				ret = append(ret, name)
				ret = append(ret, b.quote(v))
			}
		}
	}
	for _, arg := range args2 {
		if arg.IsOption() || !arg.IsProvided() {
			continue
		}
		for _, v := range arg.Value() {
			ret = append(ret, b.quote(v))
		}
	}
	return ret, nil
}

func (b *Builder) quote(s string) string {
	if b.QuoteValue && strings.ContainsRune(s, ' ') {
		s = strconv.Quote(s)
	}
	return s
}

func stringOrDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
