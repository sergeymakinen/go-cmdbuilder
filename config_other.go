// +build !windows

package cmdbuilder

import "github.com/sergeymakinen/go-quote/unix"

var defaultConfig = &Config{
	ShortOptionDelimiter:            "-",
	LongOptionDelimiter:             "--",
	OptionArgumentDelimiter:         " ",
	OptionOptionalArgumentDelimiter: "=",
	ArgumentQuoter: func(s string) string {
		if unix.DoubleQuote.MustQuote(s) {
			return unix.DoubleQuote.Quote(s)
		}
		return s
	},
}
