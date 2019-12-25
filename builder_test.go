package cmdbuilder

import (
	"reflect"
	"testing"
)

type bldOpts struct {
	DisableShortOption           bool
	DisableCombiningShortOptions bool
	ShortOptionDelimiter         string
	LongOptionDelimiter          string
	NameValueDelimiter           string
	QuoteValue                   bool
}

type bldTest struct {
	Name         string
	Opts         bldOpts
	Args         []Arg
	ExpectedArgs []string
}

type argImplTest struct {
	arg argTest
}

func (a argImplTest) IsOption() bool        { return a.arg.IsOption }
func (a argImplTest) IsProvided() bool      { return a.arg.IsProvided }
func (a argImplTest) IsValueOptional() bool { return a.arg.IsValueOptional }
func (a argImplTest) IsValueProvided() bool { return a.arg.IsValueProvided }
func (a argImplTest) Name() string          { return a.arg.Name }
func (a argImplTest) ShortName() string     { return a.arg.ShortName }
func (a argImplTest) Value() []string       { return a.arg.Value }

func TestBuilder(t *testing.T) {
	bldArgs := []Arg{
		argImplTest{argTest{
			IsOption:        true,
			IsProvided:      false,
			IsValueOptional: false,
			IsValueProvided: false,
			Name:            "unprovided",
			ShortName:       "u",
			Value:           nil,
		}},
		argImplTest{argTest{
			IsOption:        true,
			IsProvided:      true,
			IsValueOptional: true,
			IsValueProvided: false,
			Name:            "bool-true",
			ShortName:       "b",
			Value:           []string{"true"},
		}},
		argImplTest{argTest{
			IsOption:        true,
			IsProvided:      true,
			IsValueOptional: true,
			IsValueProvided: true,
			Name:            "bool-false",
			ShortName:       "B",
			Value:           []string{"false"},
		}},
		argImplTest{argTest{
			IsOption:        true,
			IsProvided:      true,
			IsValueOptional: false,
			IsValueProvided: true,
			Name:            "int",
			ShortName:       "i",
			Value:           []string{"1"},
		}},
		argImplTest{argTest{
			IsOption:        true,
			IsProvided:      true,
			IsValueOptional: true,
			IsValueProvided: false,
			Name:            "string-value-unprovided",
			ShortName:       "f",
			Value:           []string{"foo bar"},
		}},
		argImplTest{argTest{
			IsOption:        true,
			IsProvided:      true,
			IsValueOptional: true,
			IsValueProvided: true,
			Name:            "string-value-provided",
			ShortName:       "q",
			Value:           []string{"baz qux"},
		}},
		argImplTest{argTest{
			IsOption:        true,
			IsProvided:      true,
			IsValueOptional: true,
			IsValueProvided: true,
			Name:            "string-optional",
			ShortName:       "",
			Value:           []string{"foo", "bar"},
		}},
		argImplTest{argTest{
			IsOption:        false,
			IsProvided:      true,
			IsValueOptional: false,
			IsValueProvided: true,
			Name:            "",
			ShortName:       "",
			Value:           []string{"foo bar"},
		}},
	}
	bldTests := []bldTest{
		{
			Name:         "empty",
			Args:         nil,
			ExpectedArgs: nil,
		},
		{
			Name: "none",
			Args: []Arg{
				argImplTest{argTest{
					IsOption:        true,
					IsProvided:      false,
					IsValueOptional: false,
					IsValueProvided: false,
					Name:            "unprovided",
					ShortName:       "u",
					Value:           nil,
				}},
			},
			ExpectedArgs: nil,
		},
		{
			Name:         "default",
			Args:         bldArgs,
			ExpectedArgs: []string{"-bf", "-B=false", "-i", "1", "-q=baz qux", "--string-optional=foo", "--string-optional=bar", "foo bar"},
		},
		{
			Name: "DisableShortOption",
			Opts: bldOpts{
				DisableShortOption: true,
			},
			Args:         bldArgs,
			ExpectedArgs: []string{"--bool-true", "--bool-false=false", "--int", "1", "--string-value-unprovided", "--string-value-provided=baz qux", "--string-optional=foo", "--string-optional=bar", "foo bar"},
		},
		{
			Name: "DisableCombiningShortOptions",
			Opts: bldOpts{
				DisableCombiningShortOptions: true,
			},
			Args:         bldArgs,
			ExpectedArgs: []string{"-b", "-B=false", "-i", "1", "-f", "-q=baz qux", "--string-optional=foo", "--string-optional=bar", "foo bar"},
		},
		{
			Name: "ShortOptionDelimiter",
			Opts: bldOpts{
				ShortOptionDelimiter: "/",
			},
			Args:         bldArgs,
			ExpectedArgs: []string{"/bf", "/B=false", "/i", "1", "/q=baz qux", "--string-optional=foo", "--string-optional=bar", "foo bar"},
		},
		{
			Name: "LongOptionDelimiter",
			Opts: bldOpts{
				LongOptionDelimiter: "/",
			},
			Args:         bldArgs,
			ExpectedArgs: []string{"-bf", "-B=false", "-i", "1", "-q=baz qux", "/string-optional=foo", "/string-optional=bar", "foo bar"},
		},
		{
			Name: "NameValueDelimiter",
			Opts: bldOpts{
				NameValueDelimiter: ":",
			},
			Args:         bldArgs,
			ExpectedArgs: []string{"-bf", "-B:false", "-i", "1", "-q:baz qux", "--string-optional:foo", "--string-optional:bar", "foo bar"},
		},
		{
			Name: "QuoteValue",
			Opts: bldOpts{
				QuoteValue: true,
			},
			Args:         bldArgs,
			ExpectedArgs: []string{"-bf", "-B=false", "-i", "1", "-q=\"baz qux\"", "--string-optional=foo", "--string-optional=bar", "\"foo bar\""},
		},
		{
			Name: "windows",
			Opts: bldOpts{
				ShortOptionDelimiter: "/",
				LongOptionDelimiter:  "/",
				NameValueDelimiter:   ":",
			},
			Args:         bldArgs,
			ExpectedArgs: []string{"/bf", "/B:false", "/i", "1", "/q:baz qux", "/string-optional:foo", "/string-optional:bar", "foo bar"},
		},
	}
	for _, bt := range bldTests {
		t.Run("Name="+bt.Name, func(t *testing.T) {
			builder := &Builder{}
			builder.DisableShortOption = bt.Opts.DisableShortOption
			builder.DisableCombiningShortOptions = bt.Opts.DisableCombiningShortOptions
			builder.ShortOptionDelimiter = bt.Opts.ShortOptionDelimiter
			builder.LongOptionDelimiter = bt.Opts.LongOptionDelimiter
			builder.NameValueDelimiter = bt.Opts.NameValueDelimiter
			builder.QuoteValue = bt.Opts.QuoteValue
			args, err := builder.Build(bt.Args)
			if err != nil {
				t.Fatalf("Builder.Build() = _, %v", err)
			}
			if !reflect.DeepEqual(args, bt.ExpectedArgs) {
				t.Errorf("args = %v; want %v", args, bt.ExpectedArgs)
			}
		})
	}
}
