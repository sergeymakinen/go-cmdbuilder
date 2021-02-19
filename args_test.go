package cmdbuilder

import (
	"reflect"
	"strings"
	"testing"

	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
)

var testConfig = &Config{
	ShortOptionDelimiter:            "-",
	LongOptionDelimiter:             "--",
	OptionArgumentDelimiter:         " ",
	OptionOptionalArgumentDelimiter: "=",
}

func TestArgsShouldFailOnNonStructs(t *testing.T) {
	if _, err := testConfig.Args(nil); err == nil {
		t.Error("Config.Args() = _, nil; want non-nil")
	}
	if _, err := testConfig.Args(true); err == nil {
		t.Error("Config.Args() = _, nil; want non-nil")
	}
	badStruct := struct {
		Bad string `malformed`
	}{}
	if _, err := testConfig.Args(badStruct); err == nil {
		t.Error("Config.Args() = _, nil; want non-nil")
	}
	badInnerStruct := struct {
		Inner struct {
			Bad string `malformed`
		}
	}{}
	if _, err := testConfig.Args(badInnerStruct); err == nil {
		t.Error("Config.Args() = _, nil; want non-nil")
	}
}

func TestArgsShouldSkipSomeFields(t *testing.T) {
	skipStruct := struct {
		NoName  string
		Skipped string `no-flag:"true"`
		OK      string `long:"ok"`
	}{
		NoName:  "foo",
		Skipped: "foo",
		OK:      "foo",
	}
	args, err := testConfig.Args(skipStruct)
	testArgsAreEqual(t, []string{"--ok", "foo"}, args, err)
}

func TestArgsWithMaps(t *testing.T) {
	mapStruct := struct {
		Map map[string]string `long:"map"`
	}{
		Map: map[string]string{"foo": "bar1", "baz": "bar2"},
	}
	args, err := testConfig.Args(mapStruct)
	testArgsAreEqual(t, []string{"--map", "baz:bar2", "--map", "foo:bar1"}, args, err)
}

func TestArgsWithPointers(t *testing.T) {
	bt1 := true
	bt2 := &bt1
	bt3 := &bt2
	st1 := []**bool{bt3}
	st2 := &st1

	bf1 := false
	bf2 := &bf1
	bf3 := &bf2
	sf1 := []**bool{bf3}
	sf2 := &sf1

	type embeddedStruct struct {
		Uninited **[]**bool `short:"u"`
	}
	type ignoredStruct struct {
		ignored **[]**bool `short:"u"`
	}
	ptrStruct := struct {
		InitedTrue  **[]**bool `short:"i"`
		InitedFalse **[]**bool `long:"inited-false" default:"true"`
		embeddedStruct

		ignoredStruct
		ignored1 embeddedStruct
		ignored2 ignoredStruct
	}{
		InitedTrue:  &st2,
		InitedFalse: &sf2,
	}
	args, err := testConfig.Args(ptrStruct)
	testArgsAreEqual(t, []string{"-i", "--inited-false=false"}, args, err)
}

type marshalTest string

func (m marshalTest) MarshalFlag() (string, error) {
	if m == "fail" {
		return string("fail: " + m), errors.New("error")
	}
	return string("success: " + m), nil
}

func TestArgsWithMarshaler(t *testing.T) {
	s := struct {
		Value marshalTest `long:"value"`
	}{
		Value: "ok",
	}
	args, err := testConfig.Args(s)
	testArgsAreEqual(t, []string{"--value", "success: ok"}, args, err)
}

func TestArgsShouldFailOnMarshalerError(t *testing.T) {
	s := struct {
		Value marshalTest `long:"value"`
	}{
		Value: "fail",
	}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("recover() = nil; want non-nil")
		}
	}()
	testConfig.Args(s)
}

func TestArgsWithPositionalArgs(t *testing.T) {
	posTests := []struct {
		Name         string
		Args         []string
		ExpectedArgs []string
	}{
		{
			Name:         "no args",
			Args:         nil,
			ExpectedArgs: nil,
		},
		{
			Name:         "1 arg",
			Args:         []string{"foo"},
			ExpectedArgs: []string{"foo"},
		},
		{
			Name:         "2 args",
			Args:         []string{"foo", "bar"},
			ExpectedArgs: []string{"foo", "bar"},
		},
		{
			Name:         "2 args terminated",
			Args:         []string{"foo", "bar", "--", "baz"},
			ExpectedArgs: []string{"foo", "bar", "baz"},
		},
		{
			Name:         "2 args terminated with option",
			Args:         []string{"foo", "bar", "--", "--option=value"},
			ExpectedArgs: []string{"foo", "bar", "--option=value"},
		},
	}
	for _, pt := range posTests {
		t.Run(pt.Name, func(t *testing.T) {
			s := struct {
				Positional struct {
					Value1, Value2 string
					RenamedValue3  string `positional-arg-name:"Value3"`
				} `positional-args:"true"`
			}{}
			if _, err := flags.ParseArgs(&s, pt.Args); err != nil {
				t.Fatalf("flags.ParseArgs() = _, %v; want nil", err)
			}
			args, err := testConfig.Args(s)
			testArgsAreEqual(t, pt.ExpectedArgs, args, err)
		})
	}
}

type cfgTest struct {
	Name         string
	Config       Config
	Struct       interface{}
	ExpectedArgs []string
	ExpectedCmd  string
	Err          string
}

func TestArgs_CommandLine_WithConfig(t *testing.T) {
	s := struct {
		Unprovided            bool     `long:"unprovided" short:"u"`
		BoolTrue1             bool     `long:"bool-true1" short:"b"`
		BoolTrue2             bool     `long:"bool-true2" short:"c"`
		BoolFalse             bool     `long:"bool-false" short:"B" default:"true"`
		Int                   int      `long:"int" short:"i"`
		IntOptional           int      `long:"int-optional" optional:"true"`
		StringValueUnprovided string   `long:"string-value-unprovided" short:"f" optional-value:"foo bar"`
		StringValueProvided   string   `long:"string-value-provided" short:"q"`
		StringOptional        []string `long:"string-optional" optional:"true"`
		Positional            struct {
			Arg string
		} `positional-args:"true"`
	}{
		BoolTrue1:           true,
		BoolTrue2:           true,
		BoolFalse:           false,
		Int:                 1,
		IntOptional:         1,
		StringValueProvided: "baz qux",
		StringOptional:      []string{"foo", "bar"},
		Positional: struct {
			Arg string
		}{
			Arg: "foo bar",
		},
	}
	cfgTests := []cfgTest{
		{
			Name: "DisableShortName",
			Config: Config{
				DisableShortName: true,
			},
			Struct:       s,
			ExpectedArgs: []string{"--bool-true1", "--bool-true2", "--bool-false=false", "--int", "1", "--int-optional=1", "--string-value-provided", "baz qux", "--string-optional=foo", "--string-optional=bar", "foo bar"},
			ExpectedCmd:  "--bool-true1 --bool-true2 --bool-false=false --int 1 --int-optional=1 --string-value-provided baz qux --string-optional=foo --string-optional=bar foo bar",
		},
		{
			Name: "DisableShortName without long name",
			Config: Config{
				DisableShortName: true,
			},
			Struct: struct {
				Bool bool `short:"b"`
			}{
				Bool: true,
			},
			Err: "option does not have long name",
		},
		{
			Name: "DisableCombiningShortOptions",
			Config: Config{
				DisableCombiningShortOptions: true,
			},
			Struct:       s,
			ExpectedArgs: []string{"-b", "-c", "--bool-false=false", "-i", "1", "--int-optional=1", "-q", "baz qux", "--string-optional=foo", "--string-optional=bar", "foo bar"},
			ExpectedCmd:  "-b -c --bool-false=false -i 1 --int-optional=1 -q baz qux --string-optional=foo --string-optional=bar foo bar",
		},
		{
			Name: "ShortOptionDelimiter",
			Config: Config{
				ShortOptionDelimiter: "/",
			},
			Struct:       s,
			ExpectedArgs: []string{"/bc", "--bool-false=false", "/i", "1", "--int-optional=1", "/q", "baz qux", "--string-optional=foo", "--string-optional=bar", "foo bar"},
			ExpectedCmd:  "/bc --bool-false=false /i 1 --int-optional=1 /q baz qux --string-optional=foo --string-optional=bar foo bar",
		},
		{
			Name: "LongOptionDelimiter",
			Config: Config{
				LongOptionDelimiter: "/",
			},
			Struct:       s,
			ExpectedArgs: []string{"-bc", "/bool-false=false", "-i", "1", "/int-optional=1", "-q", "baz qux", "/string-optional=foo", "/string-optional=bar", "foo bar"},
			ExpectedCmd:  "-bc /bool-false=false -i 1 /int-optional=1 -q baz qux /string-optional=foo /string-optional=bar foo bar",
		},
		{
			Name: "OptionOptionalArgumentDelimiter=:",
			Config: Config{
				OptionOptionalArgumentDelimiter: ":",
			},
			Struct:       s,
			ExpectedArgs: []string{"-bc", "--bool-false:false", "-i", "1", "--int-optional:1", "-q", "baz qux", "--string-optional:foo", "--string-optional:bar", "foo bar"},
			ExpectedCmd:  "-bc --bool-false:false -i 1 --int-optional:1 -q baz qux --string-optional:foo --string-optional:bar foo bar",
		},
		{
			Name: "OptionOptionalArgumentDelimiter= ",
			Config: Config{
				OptionOptionalArgumentDelimiter: " ",
			},
			Struct:       s,
			ExpectedArgs: []string{"-bc", "--bool-false", "false", "-i", "1", "--int-optional", "1", "-q", "baz qux", "--string-optional", "foo", "--string-optional", "bar", "foo bar"},
			ExpectedCmd:  "-bc --bool-false false -i 1 --int-optional 1 -q baz qux --string-optional foo --string-optional bar foo bar",
		},
		{
			Name: "ArgumentQuoter",
			Config: Config{
				ShortOptionDelimiter:            "-",
				LongOptionDelimiter:             "--",
				OptionOptionalArgumentDelimiter: "=",
				ArgumentQuoter: func(s string) string {
					return `"` + s + `"`
				},
			},
			Struct:       s,
			ExpectedArgs: []string{"-bc", `--bool-false="false"`, "-i", "1", `--int-optional="1"`, "-q", "baz qux", `--string-optional="foo"`, `--string-optional="bar"`, "foo bar"},
			ExpectedCmd:  `-bc --bool-false="false" -i "1" --int-optional="1" -q "baz qux" --string-optional="foo" --string-optional="bar" "foo bar"`,
		},
	}
	for _, ct := range cfgTests {
		t.Run(ct.Name, func(t *testing.T) {
			config := *testConfig
			if ct.Config.DisableShortName {
				config.DisableShortName = ct.Config.DisableShortName
			}
			if ct.Config.DisableCombiningShortOptions {
				config.DisableCombiningShortOptions = ct.Config.DisableCombiningShortOptions
			}
			if ct.Config.ShortOptionDelimiter != "" {
				config.ShortOptionDelimiter = ct.Config.ShortOptionDelimiter
			}
			if ct.Config.LongOptionDelimiter != "" {
				config.LongOptionDelimiter = ct.Config.LongOptionDelimiter
			}
			if ct.Config.OptionArgumentDelimiter != "" {
				config.OptionArgumentDelimiter = ct.Config.OptionArgumentDelimiter
			}
			if ct.Config.OptionOptionalArgumentDelimiter != "" {
				config.OptionOptionalArgumentDelimiter = ct.Config.OptionOptionalArgumentDelimiter
			}
			if ct.Config.OptionsTerminator != "" {
				config.OptionsTerminator = ct.Config.OptionsTerminator
			}
			if ct.Config.ArgumentQuoter != nil {
				config.ArgumentQuoter = ct.Config.ArgumentQuoter
			}
			args, err := config.Args(ct.Struct)
			if ct.Err != "" {
				if err == nil || !strings.Contains(err.Error(), ct.Err) {
					t.Errorf("Config.Args() = _, %v; does not contain %q", err, ct.Err)
				}
			} else {
				testArgsAreEqual(t, ct.ExpectedArgs, args, err)
			}
			cmd, err := config.CommandLine(ct.Struct)
			if ct.Err != "" {
				if err == nil || !strings.Contains(err.Error(), ct.Err) {
					t.Errorf("Config.CommandLine() = _, %v; does not contain %q", err, ct.Err)
				}
			} else {
				testCmdLineIsEqual(t, ct.ExpectedCmd, cmd, err)
			}
		})
	}
}

func testArgsAreEqual(t *testing.T, expected, args []string, err error) {
	if err != nil {
		t.Fatalf("Config.Args() = _, %v; want nil", err)
	}
	if !reflect.DeepEqual(args, expected) {
		t.Errorf("Config.Args() = %v, _; want %v", args, expected)
	}
}

func testCmdLineIsEqual(t *testing.T, expected, cmdLine string, err error) {
	if err != nil {
		t.Fatalf("Config.CommandLine() = _, %v; want nil", err)
	}
	if cmdLine != expected {
		t.Errorf("Config.CommandLine() = %v, _; want %v", cmdLine, expected)
	}
}
