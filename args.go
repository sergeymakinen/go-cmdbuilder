/*
Package cmdbuilder implements a converter from structs defining command-line arguments
and their values back to a command-line.

Usage

The cmdbuilder package aims to be compatible with the flags package (see https://github.com/jessevdk/go-flags),
so it also uses structs, the reflection and struct field tags to specify command-line arguments. For example:

	type Options struct {
		Verbose    []bool            `short:"v" long:"verbose"`
		AuthorInfo map[string]string `short:"a"`
		Name       string            `long:"name" optional:"true"`
	}

This specifies the 'Verbose' boolean option with a short name '-v' and a long name '--verbose',
the 'AuthorInfo' map option with a short name '-a',
and the 'Name' string option with a long name '--name' and an optional value.
If the struct is initialized as the following:

	opts := Options{
		Verbose:    []bool{true, true, true},
		AuthorInfo: map[string]string{"name": "Jesse", "surname": "van den Kieboom"},
		Name:       "Sergey Makinen",
	}

Then the CommandLine function will produce the following string:

	-vvv -a name:Jesse -a "surname:van den Kieboom" --name="Sergey Makinen"

Any type that implements the Marshaler interface may fully customize its value output.


Arguments, options and conventions

The terms "argument" and "option" are used here as specified by Program Argument Syntax Conventions
of The GNU C Library Reference Manual (see https://www.gnu.org/software/libc/manual/html_node/Argument-Syntax.html):

 - Arguments are options if they begin with a hyphen delimiter ('-').
 - Multiple options may follow a hyphen delimiter in a single token if the options do not take arguments.
   Thus, '-abc' is equivalent to '-a -b -c'.
 - Option names are single alphanumeric characters.
 - Certain options require an argument.
 - An option and its argument may or may not appear as separate tokens. (In other words, the whitespace
   separating them is optional.) Thus, '-o foo' and '-ofoo' are equivalent.
 - Options typically precede other non-option arguments.
 - The argument '--' terminates all options; any following arguments are treated as non-option arguments,
   even if they begin with a hyphen.
 - A token consisting of a single hyphen character is interpreted as an ordinary non-option argument.
   By convention, it is used to specify input from or output to the standard input and output streams.
 - Options may be supplied in any order, or appear multiple times.
 - Long options consist of '--' followed by a name made of alphanumeric characters and dashes.
 - To specify an argument for a long option, write '--name=value'. This syntax enables a long option
   to accept an argument that is itself optional.

Supported field tags

The following is a list of supported struct field tags (which is a subset of supported tags by the flags package):

At least one is required:

    short:               the short name of the option (single character)

    long:                the long name of the option

Optional:

    no-flag:             if non-empty, this field is ignored

    optional:            if non-empty, makes the value of the option optional.
                         When the value is optional, it will produce '--argument=value'

    optional-value:      the value of the option when the argument has a zero value
                         This tag can be specified multiple times in case of maps or slices

    default:             the default value of the argument. This tag can be specified multiple times
                         in case of slices or maps

    positional-args:     when specified on a field with a struct type, uses the fields of that struct
                         (in order of the fields) as positional arguments
*/
package cmdbuilder

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
)

// FieldError represents an error when converting struct fields.
type FieldError struct {
	Struct reflect.Type // type of the struct containing the field
	Field  string       // name of the field
	Type   reflect.Type // type of the field
	Msg    string       // description of error
}

func (e *FieldError) Error() string {
	return fmt.Sprintf("failed to convert struct field %s.%s of type %s: %s", e.Struct, e.Field, e.Type, e.Msg)
}

// Quoter returns s quoted such that it appears correctly
// as a single command-line argument.
type Quoter func(s string) string

// Config controls the output of Args and CommandLine.
type Config struct {
	// DisableShortName specifies whether to use only a long name,
	// if an option has both short and long names specified.
	DisableShortName bool

	// DisableCombiningShortOptions specifies whether to not combine
	// multiple short options to an one command-line argument.
	//
	// Thus, '-a -b -c' won't become '-abc'.
	DisableCombiningShortOptions bool

	// ShortOptionDelimiter defines the delimiter is written
	// before a short option name.
	ShortOptionDelimiter string

	// WithLongOptionDelimiter defines the delimiter is written
	// before a long option name.
	LongOptionDelimiter string

	// OptionArgumentDelimiter defines the delimiter is written
	// between an option and its argument.
	//
	// Thus, the option 'o' with the argument 'foo' will become '-o foo', if the delimiter is space.
	OptionArgumentDelimiter string

	// OptionOptionalArgumentDelimiter defines the delimiter is written
	// between an option and its optional argument.
	//
	// Thus, the option 'opt' with the argument 'foo' will become '--opt=foo', if the delimiter is '='.
	OptionOptionalArgumentDelimiter string

	// OptionsTerminator defines the terminator is written
	// between options and positional arguments.
	//
	// Thus, '--a bc de -f' will become '-f --a -- bc de', if the terminator is '--'.
	OptionsTerminator string

	// ArgumentQuoter, if not nil, is applied to option and positional arguments.
	ArgumentQuoter Quoter
}

// Args converts the provided struct (or pointer to a struct) v
// defining command-line options and their values to command-line arguments
// using the provided configuration c.
func (c *Config) Args(v interface{}) ([]string, error) {
	return c.args(v, false)
}

// CommandLine converts the provided struct (or pointer to a struct) v
// defining command-line options and their values to a command-line
// using the provided configuration c.
func (c *Config) CommandLine(v interface{}) (string, error) {
	args, err := c.args(v, true)
	if err != nil {
		return "", err
	}
	return strings.Join(args, " "), nil
}

func (c *Config) args(v interface{}, cmdline bool) ([]string, error) {
	parsed, err := parse(v)
	if err != nil {
		return nil, err
	}
	parsed, args := c.combineShorts(parsed)
	var buf bytes.Buffer
	for _, arg := range parsed {
		if !arg.IsOption() || !arg.IsProvided() {
			continue
		}
		buf.Reset()
		if c.DisableShortName || arg.ShortName() == "" || (arg.IsValueOptional() && arg.IsValueProvided()) {
			buf.WriteString(c.LongOptionDelimiter)
			if arg.Name() == "" {
				return nil, &FieldError{
					Struct: arg.Struct(),
					Field:  arg.Field().Name,
					Type:   arg.Field().Type,
					Msg:    "option does not have long name",
				}
			}
			buf.WriteString(arg.Name())
			if arg.IsValueOptional() && arg.IsValueProvided() && c.OptionOptionalArgumentDelimiter != " " {
				buf.WriteString(c.OptionOptionalArgumentDelimiter)
			}
		} else {
			buf.WriteString(c.ShortOptionDelimiter)
			buf.WriteString(arg.ShortName())
		}
		for _, v := range arg.Value() {
			if arg.IsValueOptional() {
				if arg.IsValueProvided() {
					if c.OptionOptionalArgumentDelimiter == " " {
						args = append(args, buf.String())
						if cmdline {
							args = append(args, c.quote(v))
						} else {
							args = append(args, v)
						}
					} else {
						args = append(args, buf.String()+c.quote(v))
					}
				} else {
					args = append(args, buf.String())
					if !c.DisableCombiningShortOptions {
						break
					}
				}
			} else {
				args = append(args, buf.String())
				if cmdline {
					args = append(args, c.quote(v))
				} else {
					args = append(args, v)
				}
			}
		}
	}
	for _, arg := range parsed {
		if arg.IsOption() || !arg.IsProvided() {
			continue
		}
		for _, v := range arg.Value() {
			if cmdline {
				args = append(args, c.quote(v))
			} else {
				args = append(args, v)
			}
		}
	}
	return args, nil
}

func (c *Config) combineShorts(parsed []arg) (rem []arg, args []string) {
	if c.DisableCombiningShortOptions || c.DisableShortName {
		return parsed, nil
	}
	var shorts []string
	for _, arg := range parsed {
		if arg.IsOption() && arg.IsProvided() && arg.IsValueOptional() && !arg.IsValueProvided() && arg.ShortName() != "" {
			for range arg.Value() {
				shorts = append(shorts, arg.ShortName())
			}
		} else {
			rem = append(rem, arg)
		}
	}
	if len(shorts) > 0 {
		args = append(args, c.ShortOptionDelimiter+strings.Join(shorts, ""))
	}
	return
}

func (c *Config) quote(s string) string {
	if c.ArgumentQuoter != nil {
		return c.ArgumentQuoter(s)
	}
	return s
}

// Args converts the provided struct (or pointer to a struct) v
// defining command-line options and their values to command-line arguments
// using a default configuration.
func Args(v interface{}) ([]string, error) {
	return defaultConfig.Args(v)
}

// CommandLine converts the provided struct (or pointer to a struct) v
// defining command-line options and their values to a command-line
// using a default configuration.
func CommandLine(v interface{}) (string, error) {
	return defaultConfig.CommandLine(v)
}
