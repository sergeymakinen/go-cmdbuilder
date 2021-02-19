package cmdbuilder

import (
	"bytes"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/ompluscator/dynamic-struct"
)

type optTest struct {
	IsSlice         bool
	Args            []string
	Name            string
	ShortName       string
	Default         interface{}
	OptionalDefault []string
	ExpectedArgs    []string
}

var optTests = []optTest{
	// bool
	{
		Args:            nil,
		Name:            "unprovided-bool",
		ShortName:       "",
		Default:         false,
		OptionalDefault: nil,
		ExpectedArgs:    nil,
	},
	{
		Args:            []string{"--provided-bool"},
		Name:            "provided-bool",
		ShortName:       "",
		Default:         false,
		OptionalDefault: nil,
		ExpectedArgs:    []string{"--provided-bool"},
	},
	{
		Args:            []string{"--provided-bool-zero-default-non-zero-value=true"},
		Name:            "provided-bool-zero-default-non-zero-value",
		ShortName:       "",
		Default:         false,
		OptionalDefault: nil,
		ExpectedArgs:    []string{"--provided-bool-zero-default-non-zero-value"},
	},
	{
		Args:            []string{"--provided-bool-zero-default-zero-value=false"},
		Name:            "provided-bool-zero-default-zero-value",
		ShortName:       "",
		Default:         false,
		OptionalDefault: nil,
		ExpectedArgs:    nil,
	},
	{
		Args:            nil,
		Name:            "unprovided-bool-non-zero-default",
		ShortName:       "",
		Default:         true,
		OptionalDefault: nil,
		ExpectedArgs:    nil,
	},
	{
		Args:            []string{"--provided-bool-non-zero-default"},
		Name:            "provided-bool-non-zero-default",
		ShortName:       "",
		Default:         true,
		OptionalDefault: nil,
		ExpectedArgs:    nil,
	},
	{
		Args:            []string{"--provided-bool-non-zero-default-non-zero-value=true"},
		Name:            "provided-bool-non-zero-default-non-zero-value",
		ShortName:       "",
		Default:         true,
		OptionalDefault: nil,
		ExpectedArgs:    nil,
	},
	{
		Args:            []string{"--provided-bool-non-zero-default-zero-value=false"},
		Name:            "provided-bool-non-zero-default-zero-value",
		ShortName:       "",
		Default:         true,
		OptionalDefault: nil,
		ExpectedArgs:    []string{"--provided-bool-non-zero-default-zero-value=false"},
	},

	// []bool
	{
		IsSlice:         true,
		Args:            nil,
		Name:            "unprovided-bool-slice",
		ShortName:       "",
		Default:         []bool{false, false},
		OptionalDefault: nil,
		ExpectedArgs:    nil,
	},
	{
		IsSlice:         true,
		Args:            nil,
		Name:            "unprovided-bool-slice-optional",
		ShortName:       "",
		Default:         []bool{false, false},
		OptionalDefault: []string{"false", "true"},
		ExpectedArgs:    nil,
	},
	{
		IsSlice:         true,
		Args:            []string{"--provided-bool-slice-1-non-zero=true"},
		Name:            "provided-bool-slice-1-non-zero",
		ShortName:       "",
		Default:         []bool{false, false},
		OptionalDefault: nil,
		ExpectedArgs:    []string{"--provided-bool-slice-1-non-zero"},
	},
	{
		IsSlice:         true,
		Args:            []string{"--provided-bool-slice-optional-1-non-zero=true"},
		Name:            "provided-bool-slice-optional-1-non-zero",
		ShortName:       "",
		Default:         []bool{false, false},
		OptionalDefault: []string{"false", "true"},
		ExpectedArgs:    []string{"--provided-bool-slice-optional-1-non-zero"},
	},
	{
		IsSlice:         true,
		Args:            []string{"--provided-bool-slice-2-non-zero=true", "--provided-bool-slice-2-non-zero=true"},
		Name:            "provided-bool-slice-2-non-zero",
		ShortName:       "",
		Default:         []bool{false, false},
		OptionalDefault: nil,
		ExpectedArgs:    []string{"--provided-bool-slice-2-non-zero", "--provided-bool-slice-2-non-zero"},
	},
	{
		IsSlice:         true,
		Args:            []string{"--provided-bool-slice-optional-2-non-zero=true", "--provided-bool-slice-optional-2-non-zero=true"},
		Name:            "provided-bool-slice-optional-2-non-zero",
		ShortName:       "",
		Default:         []bool{false, false},
		OptionalDefault: []string{"false", "true"},
		ExpectedArgs:    []string{"--provided-bool-slice-optional-2-non-zero", "--provided-bool-slice-optional-2-non-zero"},
	},
	{
		IsSlice:         true,
		Args:            []string{"--provided-bool-slice-1-non-zero-1-zero=true", "--provided-bool-slice-1-non-zero-1-zero=false"},
		Name:            "provided-bool-slice-1-non-zero-1-zero",
		ShortName:       "",
		Default:         []bool{false, false},
		OptionalDefault: nil,
		ExpectedArgs:    []string{"--provided-bool-slice-1-non-zero-1-zero", "--provided-bool-slice-1-non-zero-1-zero=false"},
	},
	{
		IsSlice:         true,
		Args:            []string{"--provided-bool-slice-optional-1-non-zero-1-zero=true", "--provided-bool-slice-optional-1-non-zero-1-zero=false"},
		Name:            "provided-bool-slice-optional-1-non-zero-1-zero",
		ShortName:       "",
		Default:         []bool{false, false},
		OptionalDefault: []string{"false", "true"},
		ExpectedArgs:    []string{"--provided-bool-slice-optional-1-non-zero-1-zero", "--provided-bool-slice-optional-1-non-zero-1-zero=false"},
	},
	{
		IsSlice:         true,
		Args:            []string{"--provided-bool-slice-optional"},
		Name:            "provided-bool-slice-optional",
		ShortName:       "",
		Default:         []bool{false, false},
		OptionalDefault: []string{"false", "true"},
		ExpectedArgs:    []string{"--provided-bool-slice-optional"},
	},
}

func TestArgsWithOptions(t *testing.T) {
	for _, ot := range optTests {
		t.Run(ot.Name, func(t *testing.T) {
			tag := `short:"` + ot.ShortName + `" long:"` + ot.Name + `"`
			if ot.IsSlice {
				isZeroBool := false
				if bb, ok := ot.Default.([]bool); ok {
					isZeroBool = true
					for _, b := range bb {
						if b {
							isZeroBool = false
							break
						}
					}
				}
				if !isZeroBool {
					v := reflect.ValueOf(ot.Default)
					for i := 0; i < v.Len(); i++ {
						tag += ` default:"` + toValue(v.Index(i).Interface())[0] + `"`
					}
				}
			} else {
				if b, ok := ot.Default.(bool); !ok || b {
					for _, v := range toValue(ot.Default) {
						tag += ` default:"` + v + `"`
					}
				}
			}
			if ot.OptionalDefault != nil {
				tag += ` optional:"true"`
				for _, v := range ot.OptionalDefault {
					tag += ` optional-value:"` + v + `"`
				}
			}
			var s interface{}
			switch ot.Default.(type) {
			case bool, []bool, float32, []float32, float64, []float64, int8, []int8, uint8, []uint8,
				int16, []int16, uint16, []uint16, int32, []int32, uint32, []uint32, int64, []int64, uint64, []uint64,
				int, []int, uint, []uint, string, []string, time.Duration, []time.Duration:
				if !strings.Contains(ot.Name, "-byte-") && !strings.HasSuffix(ot.Name, "-byte") {
					s = dynamicstruct.NewStruct().
						AddField("Value", ot.Default, tag).
						Build().
						New()
				} else {
					t.Skip("reflect.ValueOf(ot.Default).Type() = []byte")
				}
			default:
				t.Skipf("reflect.ValueOf(ot.Default).Type() = %s", reflect.ValueOf(ot.Default).Type())
			}
			if _, err := flags.ParseArgs(s, ot.Args); err != nil {
				if ferr, ok := err.(*flags.Error); ok && !isError(ferr) {
					t.Skip(err.Error())
				} else {
					t.Fatalf("flags.ParseArgs() = _, %v; want nil", err)
				}
			}
			args, err := testConfig.Args(s)
			testArgsAreEqual(t, ot.ExpectedArgs, args, err)
		})
	}
}

func toValue(args ...interface{}) []string {
	var list []string
	for _, arg := range args {
		switch arg.(type) {
		case bool:
			list = append(list, strconv.FormatBool(arg.(bool)))
		case int8, uint8, int16, uint16, int32, uint32, int64, uint64, int, uint:
			list = append(list, fmt.Sprintf("%d", arg))
		case float32, float64:
			list = append(list, fmt.Sprintf("%g", arg))
		case []byte:
			list = append(list, fmt.Sprintf("%X", arg))
		case net.IPNet:
			n := arg.(net.IPNet)
			list = append(list, fmt.Sprintf("%s", (&n).String()))
		default:
			list = append(list, fmt.Sprintf("%s", arg))
		}
	}
	return list
}

func isError(err *flags.Error) bool {
	if err.Type == flags.ErrNoArgumentForBool {
		return false
	}
	return err.Type != flags.ErrInvalidTag || !strings.HasPrefix(err.Error(), "boolean")
}

func addScalarTests(name string, zeroVal, nonZeroVal1, nonZeroVal2 interface{}) {
	newTests := []optTest{
		{
			Args:            nil,
			Name:            "unprovided-{{.Name}}",
			Default:         zeroVal,
			OptionalDefault: nil,
			ExpectedArgs:    nil,
		},
		{
			Args:            nil,
			Name:            "unprovided-{{.Name}}-non-zero-optional",
			Default:         zeroVal,
			OptionalDefault: toValue(nonZeroVal1),
			ExpectedArgs:    nil,
		},
		{
			Args:            []string{"--provided-{{.Name}}-zero-default-non-zero-value-1", "{{.NonZeroVal1}}"},
			Name:            "provided-{{.Name}}-zero-default-non-zero-value-1",
			Default:         zeroVal,
			OptionalDefault: nil,
			ExpectedArgs:    []string{"--provided-{{.Name}}-zero-default-non-zero-value-1", "{{.NonZeroVal1}}"},
		},
		{
			Args:            []string{"--provided-{{.Name}}-zero-default-non-zero-value-2={{.NonZeroVal1}}"},
			Name:            "provided-{{.Name}}-zero-default-non-zero-value-2",
			Default:         zeroVal,
			OptionalDefault: nil,
			ExpectedArgs:    []string{"--provided-{{.Name}}-zero-default-non-zero-value-2", "{{.NonZeroVal1}}"},
		},
		{
			Args:            []string{"--provided-{{.Name}}-zero-default-zero-value-1", "{{.ZeroVal}}"},
			Name:            "provided-{{.Name}}-zero-default-zero-value-1",
			Default:         zeroVal,
			OptionalDefault: nil,
			ExpectedArgs:    nil,
		},
		{
			Args:            []string{"--provided-{{.Name}}-zero-default-zero-value-2={{.ZeroVal}}"},
			Name:            "provided-{{.Name}}-zero-default-zero-value-2",
			Default:         zeroVal,
			OptionalDefault: nil,
			ExpectedArgs:    nil,
		},
		{
			Args:            []string{"--provided-{{.Name}}-non-zero-optional"},
			Name:            "provided-{{.Name}}-non-zero-optional",
			Default:         zeroVal,
			OptionalDefault: toValue(nonZeroVal1),
			ExpectedArgs:    []string{"--provided-{{.Name}}-non-zero-optional"},
		},
		{
			Args:            []string{"--provided-{{.Name}}-non-zero-optional-zero-value={{.ZeroVal}}"},
			Name:            "provided-{{.Name}}-non-zero-optional-zero-value",
			Default:         zeroVal,
			OptionalDefault: toValue(nonZeroVal1),
			ExpectedArgs:    nil,
		},
		{
			Args:            []string{"--provided-{{.Name}}-non-zero-optional-1-value={{.NonZeroVal1}}"},
			Name:            "provided-{{.Name}}-non-zero-optional-1-value",
			Default:         zeroVal,
			OptionalDefault: toValue(nonZeroVal1),
			ExpectedArgs:    []string{"--provided-{{.Name}}-non-zero-optional-1-value"},
		},
		{
			Args:            []string{"--provided-{{.Name}}-non-zero-optional-2-value={{.NonZeroVal2}}"},
			Name:            "provided-{{.Name}}-non-zero-optional-2-value",
			Default:         zeroVal,
			OptionalDefault: toValue(nonZeroVal1),
			ExpectedArgs:    []string{"--provided-{{.Name}}-non-zero-optional-2-value={{.NonZeroVal2}}"},
		},
	}
	addOptTests(false, newTests, tmplData{
		Name:        name,
		ZeroVal:     toValue(zeroVal)[0],
		NonZeroVal1: toValue(nonZeroVal1)[0],
		NonZeroVal2: toValue(nonZeroVal2)[0],
	})
	addSliceTests(name, zeroVal, nonZeroVal1)
}

type tmplData struct {
	Name, ZeroVal, NonZeroVal, NonZeroVal1, NonZeroVal2 string
}

func addOptTests(isSlice bool, tests []optTest, data tmplData) {
	replace := func(s string) string {
		t := template.Must(template.New("test").Parse(s))
		var b bytes.Buffer
		if err := t.Execute(&b, data); err != nil {
			panic(err)
		}
		return b.String()
	}
	for _, t := range tests {
		t.IsSlice = isSlice
		t.Name = replace(t.Name)
		for k, v := range t.Args {
			t.Args[k] = replace(v)
		}
		for k, v := range t.ExpectedArgs {
			t.ExpectedArgs[k] = replace(v)
		}
		optTests = append(optTests, t)
	}
}

func addSliceTests(name string, zeroVal, nonZeroVal interface{}) {
	newTests := []optTest{
		{
			Args:            nil,
			Name:            "unprovided-{{.Name}}-slice",
			Default:         toSlice(zeroVal, zeroVal),
			OptionalDefault: nil,
			ExpectedArgs:    nil,
		},
		{
			Args:            nil,
			Name:            "unprovided-{{.Name}}-slice-optional",
			Default:         toSlice(zeroVal, zeroVal),
			OptionalDefault: toValue(zeroVal, nonZeroVal),
			ExpectedArgs:    nil,
		},
		{
			Args:            []string{"--provided-{{.Name}}-slice-1-non-zero={{.NonZeroVal}}"},
			Name:            "provided-{{.Name}}-slice-1-non-zero",
			Default:         toSlice(zeroVal, zeroVal),
			OptionalDefault: nil,
			ExpectedArgs:    []string{"--provided-{{.Name}}-slice-1-non-zero", "{{.NonZeroVal}}"},
		},
		{
			Args:            []string{"--provided-{{.Name}}-slice-optional-1-non-zero={{.NonZeroVal}}"},
			Name:            "provided-{{.Name}}-slice-optional-1-non-zero",
			Default:         toSlice(zeroVal, zeroVal),
			OptionalDefault: toValue(zeroVal, nonZeroVal),
			ExpectedArgs:    []string{"--provided-{{.Name}}-slice-optional-1-non-zero={{.NonZeroVal}}"},
		},
		{
			Args:            []string{"--provided-{{.Name}}-slice-2-non-zero={{.NonZeroVal}}", "--provided-{{.Name}}-slice-2-non-zero={{.NonZeroVal}}"},
			Name:            "provided-{{.Name}}-slice-2-non-zero",
			Default:         toSlice(zeroVal, zeroVal),
			OptionalDefault: nil,
			ExpectedArgs:    []string{"--provided-{{.Name}}-slice-2-non-zero", "{{.NonZeroVal}}", "--provided-{{.Name}}-slice-2-non-zero", "{{.NonZeroVal}}"},
		},
		{
			Args:            []string{"--provided-{{.Name}}-slice-optional-2-non-zero={{.NonZeroVal}}", "--provided-{{.Name}}-slice-optional-2-non-zero={{.NonZeroVal}}"},
			Name:            "provided-{{.Name}}-slice-optional-2-non-zero",
			Default:         toSlice(zeroVal, zeroVal),
			OptionalDefault: toValue(zeroVal, nonZeroVal),
			ExpectedArgs:    []string{"--provided-{{.Name}}-slice-optional-2-non-zero={{.NonZeroVal}}", "--provided-{{.Name}}-slice-optional-2-non-zero={{.NonZeroVal}}"},
		},
		{
			Args:            []string{"--provided-{{.Name}}-slice-1-non-zero-1-zero={{.NonZeroVal}}", "--provided-{{.Name}}-slice-1-non-zero-1-zero={{.ZeroVal}}"},
			Name:            "provided-{{.Name}}-slice-1-non-zero-1-zero",
			Default:         toSlice(zeroVal, zeroVal),
			OptionalDefault: nil,
			ExpectedArgs:    []string{"--provided-{{.Name}}-slice-1-non-zero-1-zero", "{{.NonZeroVal}}", "--provided-{{.Name}}-slice-1-non-zero-1-zero", "{{.ZeroVal}}"},
		},
		{
			Args:            []string{"--provided-{{.Name}}-slice-optional-1-non-zero-1-zero={{.NonZeroVal}}", "--provided-{{.Name}}-slice-optional-1-non-zero-1-zero={{.ZeroVal}}"},
			Name:            "provided-{{.Name}}-slice-optional-1-non-zero-1-zero",
			Default:         toSlice(zeroVal, zeroVal),
			OptionalDefault: toValue(zeroVal, nonZeroVal),
			ExpectedArgs:    []string{"--provided-{{.Name}}-slice-optional-1-non-zero-1-zero={{.NonZeroVal}}", "--provided-{{.Name}}-slice-optional-1-non-zero-1-zero={{.ZeroVal}}"},
		},
		{
			Args:            []string{"--provided-{{.Name}}-slice-optional"},
			Name:            "provided-{{.Name}}-slice-optional",
			ShortName:       "",
			Default:         toSlice(zeroVal, zeroVal),
			OptionalDefault: toValue(zeroVal, nonZeroVal),
			ExpectedArgs:    []string{"--provided-{{.Name}}-slice-optional"},
		},
	}
	addOptTests(true, newTests, tmplData{
		Name:       name,
		ZeroVal:    toValue(zeroVal)[0],
		NonZeroVal: toValue(nonZeroVal)[0],
	})
}

func toSlice(args ...interface{}) interface{} {
	typ := reflect.TypeOf(args[0])
	slice := reflect.New(reflect.SliceOf(typ)).Elem()
	for _, v := range args {
		slice = reflect.Append(slice, reflect.ValueOf(v))
	}
	return slice.Interface()
}

func init() {
	addScalarTests("int8", int8(0), int8(1), int8(2))
	addScalarTests("uint8", uint8(0), uint8(1), uint8(2))
	addScalarTests("int16", int16(0), int16(1), int16(2))
	addScalarTests("uint16", uint16(0), uint16(1), uint16(2))
	addScalarTests("int32", int32(0), int32(1), int32(2))
	addScalarTests("uint32", uint32(0), uint32(1), uint32(2))
	addScalarTests("int64", int64(0), int64(1), int64(2))
	addScalarTests("uint64", uint64(0), uint64(1), uint64(2))
	addScalarTests("int", 0, 1, 2)
	addScalarTests("uint", uint(0), uint(1), uint(2))
	addScalarTests("float32", float32(0.0), float32(1.1), float32(2.2))
	addScalarTests("float64", 0.0, 1.1, 2.2)
	addScalarTests("string", "", "foo bar", "baz qux")
	// addScalarTests("byte", []byte{}, []byte{1}, []byte{1, 2}) // Hex
	addScalarTests("duration", 0*time.Second, 1*time.Minute, 2*time.Hour)
	// addScalarTests("ip", net.IPv4zero, net.IPv4(8, 8, 8, 8), net.IPv4(1, 1, 1, 1))
	// mustParseCIDR := func(s string) net.IPNet {
	// 	_, n, err := net.ParseCIDR(s)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	return *n
	// }
	// addScalarTests("ipnet", net.IPNet{
	// 	IP:   net.IPv4zero,
	// 	Mask: net.IPv4Mask(0, 0, 0, 0),
	// }, mustParseCIDR("8.8.8.8/24"), mustParseCIDR("1.1.1.1/8"))
}
