package cmdbuilder

import "github.com/sergeymakinen/go-quote/windows"

var defaultConfig = &Config{
	DisableCombiningShortOptions:    true,
	ShortOptionDelimiter:            "/",
	LongOptionDelimiter:             "/",
	OptionArgumentDelimiter:         " ",
	OptionOptionalArgumentDelimiter: ":",
	OptionsTerminator:               "",
	ArgumentQuoter: func(s string) string {
		if windows.Argv.MustQuote(s) {
			return windows.Argv.Quote(s)
		}
		return s
	},
}
