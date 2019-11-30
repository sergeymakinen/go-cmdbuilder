package cmdbuilder

import (
	"bytes"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"testing"
	"text/template"
	"time"
)

type optTest struct {
	IsSlice         bool
	Args            []string
	Name            string
	ShortName       string
	Default         interface{}
	OptionalDefault []string
	ExpectedArg     argTest
}

type posTest struct {
	Name         string
	Args         []string
	ExpectedArgs []argTest
}

type argTest struct {
	IsOption        bool
	IsProvided      bool
	IsValueOptional bool
	IsValueProvided bool
	Name            string
	ShortName       string
	Value           []string
}

func (a *argTest) TestEqual(t *testing.T, arg Arg) {
	if a.IsOption != arg.IsOption() {
		t.Errorf("Arg.IsOption() = %v; want %v", arg.IsOption(), a.IsOption)
	}
	if a.IsProvided != arg.IsProvided() {
		t.Errorf("Arg.IsProvided() = %v; want %v", arg.IsProvided(), a.IsProvided)
	}
	if a.IsValueOptional != arg.IsValueOptional() {
		t.Errorf("Arg.IsValueOptional() = %v; want %v", arg.IsValueOptional(), a.IsValueOptional)
	}
	if a.IsValueProvided != arg.IsValueProvided() {
		t.Errorf("Arg.IsValueProvided() = %v; want %v", arg.IsValueProvided(), a.IsValueProvided)
	}
	if a.IsOption && a.Name != arg.Name() {
		t.Errorf("Arg.Name() = %v; want %v", arg.Name(), a.Name)
	}
	if a.IsOption && a.ShortName != arg.ShortName() {
		t.Errorf("Arg.ShortName() = %v; want %v", arg.ShortName(), a.ShortName)
	}
	if len(a.Value) != 0 && len(arg.Value()) != 0 && !reflect.DeepEqual(a.Value, arg.Value()) {
		t.Errorf("Arg.Value() = %v; want %v", arg.Value(), a.Value)
	}
}

var optTests = []optTest{
	// bool
	{
		Args:            nil,
		Name:            "unprovided-bool",
		ShortName:       "",
		Default:         false,
		OptionalDefault: nil,
		ExpectedArg: argTest{
			IsOption:        true,
			IsProvided:      false,
			IsValueOptional: true,
			IsValueProvided: false,
			Name:            "unprovided-bool",
			ShortName:       "",
			Value:           []string{"false"},
		},
	},
	{
		Args:            []string{"--provided-bool"},
		Name:            "provided-bool",
		ShortName:       "",
		Default:         false,
		OptionalDefault: nil,
		ExpectedArg: argTest{
			IsOption:        true,
			IsProvided:      true,
			IsValueOptional: true,
			IsValueProvided: false,
			Name:            "provided-bool",
			ShortName:       "",
			Value:           []string{"true"},
		},
	},
	{
		Args:            []string{"--provided-bool-zero-default-non-zero-value=true"},
		Name:            "provided-bool-zero-default-non-zero-value",
		ShortName:       "",
		Default:         false,
		OptionalDefault: nil,
		ExpectedArg: argTest{
			IsOption:        true,
			IsProvided:      true,
			IsValueOptional: true,
			IsValueProvided: false,
			Name:            "provided-bool-zero-default-non-zero-value",
			ShortName:       "",
			Value:           []string{"true"},
		},
	},
	{
		Args:            []string{"--provided-bool-zero-default-zero-value=false"},
		Name:            "provided-bool-zero-default-zero-value",
		ShortName:       "",
		Default:         false,
		OptionalDefault: nil,
		ExpectedArg: argTest{
			IsOption:        true,
			IsProvided:      false,
			IsValueOptional: true,
			IsValueProvided: false,
			Name:            "provided-bool-zero-default-zero-value",
			ShortName:       "",
			Value:           []string{"false"},
		},
	},
	{
		Args:            nil,
		Name:            "unprovided-bool-non-zero-default",
		ShortName:       "",
		Default:         true,
		OptionalDefault: nil,
		ExpectedArg: argTest{
			IsOption:        true,
			IsProvided:      false,
			IsValueOptional: true,
			IsValueProvided: false,
			Name:            "unprovided-bool-non-zero-default",
			ShortName:       "",
			Value:           []string{"true"},
		},
	},
	{
		Args:            []string{"--provided-bool-non-zero-default"},
		Name:            "provided-bool-non-zero-default",
		ShortName:       "",
		Default:         true,
		OptionalDefault: nil,
		ExpectedArg: argTest{
			IsOption:        true,
			IsProvided:      false,
			IsValueOptional: true,
			IsValueProvided: false,
			Name:            "provided-bool-non-zero-default",
			ShortName:       "",
			Value:           []string{"true"},
		},
	},
	{
		Args:            []string{"--provided-bool-non-zero-default-non-zero-value=true"},
		Name:            "provided-bool-non-zero-default-non-zero-value",
		ShortName:       "",
		Default:         true,
		OptionalDefault: nil,
		ExpectedArg: argTest{
			IsOption:        true,
			IsProvided:      false,
			IsValueOptional: true,
			IsValueProvided: false,
			Name:            "provided-bool-non-zero-default-non-zero-value",
			ShortName:       "",
			Value:           []string{"true"},
		},
	},
	{
		Args:            []string{"--provided-bool-non-zero-default-zero-value=false"},
		Name:            "provided-bool-non-zero-default-zero-value",
		ShortName:       "",
		Default:         true,
		OptionalDefault: nil,
		ExpectedArg: argTest{
			IsOption:        true,
			IsProvided:      true,
			IsValueOptional: true,
			IsValueProvided: true,
			Name:            "provided-bool-non-zero-default-zero-value",
			ShortName:       "",
			Value:           []string{"false"},
		},
	},

	// []bool
	{
		IsSlice:         true,
		Args:            nil,
		Name:            "unprovided-bool-slice",
		ShortName:       "",
		Default:         []bool{false, false},
		OptionalDefault: nil,
		ExpectedArg: argTest{
			IsOption:        true,
			IsProvided:      false,
			IsValueOptional: true,
			IsValueProvided: false,
			Name:            "unprovided-bool-slice",
			ShortName:       "",
			Value:           []string{"false", "false"},
		},
	},
	{
		IsSlice:         true,
		Args:            nil,
		Name:            "unprovided-bool-slice-optional",
		ShortName:       "",
		Default:         []bool{false, false},
		OptionalDefault: []string{"false", "true"},
		ExpectedArg: argTest{
			IsOption:        true,
			IsProvided:      false,
			IsValueOptional: true,
			IsValueProvided: false,
			Name:            "unprovided-bool-slice-optional",
			ShortName:       "",
			Value:           []string{"false", "false"},
		},
	},
	{
		IsSlice:         true,
		Args:            []string{"--provided-bool-slice-1-non-zero=true"},
		Name:            "provided-bool-slice-1-non-zero",
		ShortName:       "",
		Default:         []bool{false, false},
		OptionalDefault: nil,
		ExpectedArg: argTest{
			IsOption:        true,
			IsProvided:      true,
			IsValueOptional: true,
			IsValueProvided: true,
			Name:            "provided-bool-slice-1-non-zero",
			ShortName:       "",
			Value:           []string{"true"},
		},
	},
	{
		IsSlice:         true,
		Args:            []string{"--provided-bool-slice-optional-1-non-zero=true"},
		Name:            "provided-bool-slice-optional-1-non-zero",
		ShortName:       "",
		Default:         []bool{false, false},
		OptionalDefault: []string{"false", "true"},
		ExpectedArg: argTest{
			IsOption:        true,
			IsProvided:      true,
			IsValueOptional: true,
			IsValueProvided: true,
			Name:            "provided-bool-slice-optional-1-non-zero",
			ShortName:       "",
			Value:           []string{"true"},
		},
	},
	{
		IsSlice:         true,
		Args:            []string{"--provided-bool-slice-2-non-zero=true", "--provided-bool-slice-2-non-zero=true"},
		Name:            "provided-bool-slice-2-non-zero",
		ShortName:       "",
		Default:         []bool{false, false},
		OptionalDefault: nil,
		ExpectedArg: argTest{
			IsOption:        true,
			IsProvided:      true,
			IsValueOptional: true,
			IsValueProvided: true,
			Name:            "provided-bool-slice-2-non-zero",
			ShortName:       "",
			Value:           []string{"true", "true"},
		},
	},
	{
		IsSlice:         true,
		Args:            []string{"--provided-bool-slice-optional-2-non-zero=true", "--provided-bool-slice-optional-2-non-zero=true"},
		Name:            "provided-bool-slice-optional-2-non-zero",
		ShortName:       "",
		Default:         []bool{false, false},
		OptionalDefault: []string{"false", "true"},
		ExpectedArg: argTest{
			IsOption:        true,
			IsProvided:      true,
			IsValueOptional: true,
			IsValueProvided: true,
			Name:            "provided-bool-slice-optional-2-non-zero",
			ShortName:       "",
			Value:           []string{"true", "true"},
		},
	},
	{
		IsSlice:         true,
		Args:            []string{"--provided-bool-slice-1-non-zero-1-zero=true", "--provided-bool-slice-1-non-zero-1-zero=false"},
		Name:            "provided-bool-slice-1-non-zero-1-zero",
		ShortName:       "",
		Default:         []bool{false, false},
		OptionalDefault: nil,
		ExpectedArg: argTest{
			IsOption:        true,
			IsProvided:      true,
			IsValueOptional: true,
			IsValueProvided: true,
			Name:            "provided-bool-slice-1-non-zero-1-zero",
			ShortName:       "",
			Value:           []string{"true", "false"},
		},
	},
	{
		IsSlice:         true,
		Args:            []string{"--provided-bool-slice-optional-1-non-zero-1-zero=true", "--provided-bool-slice-optional-1-non-zero-1-zero=false"},
		Name:            "provided-bool-slice-optional-1-non-zero-1-zero",
		ShortName:       "",
		Default:         []bool{false, false},
		OptionalDefault: []string{"false", "true"},
		ExpectedArg: argTest{
			IsOption:        true,
			IsProvided:      true,
			IsValueOptional: true,
			IsValueProvided: true,
			Name:            "provided-bool-slice-optional-1-non-zero-1-zero",
			ShortName:       "",
			Value:           []string{"true", "false"},
		},
	},
	{
		IsSlice:         true,
		Args:            []string{"--provided-bool-slice-optional"},
		Name:            "provided-bool-slice-optional",
		ShortName:       "",
		Default:         []bool{false, false},
		OptionalDefault: []string{"false", "true"},
		ExpectedArg: argTest{
			IsOption:        true,
			IsProvided:      true,
			IsValueOptional: true,
			IsValueProvided: false,
			Name:            "provided-bool-slice-optional",
			ShortName:       "",
			Value:           []string{"false", "true"},
		},
	},
}

func toSlice(args ...interface{}) interface{} {
	typ := reflect.TypeOf(args[0])
	slice := reflect.New(reflect.SliceOf(typ)).Elem()
	for _, v := range args {
		slice = reflect.Append(slice, reflect.ValueOf(v))
	}
	return slice.Interface()
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
			if arg == float32(0.0) || arg == 0.0 {
				list = append(list, "0")
			} else {
				list = append(list, fmt.Sprintf("%.1f", arg))
			}
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

func addScalarTests(name string, zeroVal, nonZeroVal1, nonZeroVal2 interface{}) {
	newTests := []optTest{
		{
			Args:            nil,
			Name:            "unprovided-{{.Name}}",
			ShortName:       "",
			Default:         zeroVal,
			OptionalDefault: nil,
			ExpectedArg: argTest{
				IsOption:        true,
				IsProvided:      false,
				IsValueOptional: false,
				IsValueProvided: false,
				Value:           toValue(zeroVal),
			},
		},
		{
			Args:            nil,
			Name:            "unprovided-{{.Name}}-non-zero-optional",
			ShortName:       "",
			Default:         zeroVal,
			OptionalDefault: toValue(nonZeroVal1),
			ExpectedArg: argTest{
				IsOption:        true,
				IsProvided:      false,
				IsValueOptional: true,
				IsValueProvided: false,
				Value:           toValue(zeroVal),
			},
		},
		{
			Args:            []string{"--provided-{{.Name}}-zero-default-non-zero-value-1", "{{.NonZeroVal1}}"},
			Name:            "provided-{{.Name}}-zero-default-non-zero-value-1",
			ShortName:       "",
			Default:         zeroVal,
			OptionalDefault: nil,
			ExpectedArg: argTest{
				IsOption:        true,
				IsProvided:      true,
				IsValueOptional: false,
				IsValueProvided: true,
				Value:           toValue(nonZeroVal1),
			},
		},
		{
			Args:            []string{"--provided-{{.Name}}-zero-default-non-zero-value-2={{.NonZeroVal1}}"},
			Name:            "provided-{{.Name}}-zero-default-non-zero-value-2",
			ShortName:       "",
			Default:         zeroVal,
			OptionalDefault: nil,
			ExpectedArg: argTest{
				IsOption:        true,
				IsProvided:      true,
				IsValueOptional: false,
				IsValueProvided: true,
				Value:           toValue(nonZeroVal1),
			},
		},
		{
			Args:            []string{"--provided-{{.Name}}-zero-default-zero-value-1", "{{.ZeroVal}}"},
			Name:            "provided-{{.Name}}-zero-default-zero-value-1",
			ShortName:       "",
			Default:         zeroVal,
			OptionalDefault: nil,
			ExpectedArg: argTest{
				IsOption:        true,
				IsProvided:      false,
				IsValueOptional: false,
				IsValueProvided: false,
				Name:            "provided-{{.Name}}-zero-default-zero-value-1",
				ShortName:       "",
				Value:           toValue(zeroVal),
			},
		},
		{
			Args:            []string{"--provided-{{.Name}}-zero-default-zero-value-2={{.ZeroVal}}"},
			Name:            "provided-{{.Name}}-zero-default-zero-value-2",
			ShortName:       "",
			Default:         zeroVal,
			OptionalDefault: nil,
			ExpectedArg: argTest{
				IsOption:        true,
				IsProvided:      false,
				IsValueOptional: false,
				IsValueProvided: false,
				Value:           toValue(zeroVal),
			},
		},
		{
			Args:            []string{"--provided-{{.Name}}-non-zero-optional"},
			Name:            "provided-{{.Name}}-non-zero-optional",
			ShortName:       "",
			Default:         zeroVal,
			OptionalDefault: toValue(nonZeroVal1),
			ExpectedArg: argTest{
				IsOption:        true,
				IsProvided:      true,
				IsValueOptional: true,
				IsValueProvided: false,
				Value:           toValue(nonZeroVal1),
			},
		},
		{
			Args:            []string{"--provided-{{.Name}}-non-zero-optional-zero-value={{.ZeroVal}}"},
			Name:            "provided-{{.Name}}-non-zero-optional-zero-value",
			ShortName:       "",
			Default:         zeroVal,
			OptionalDefault: toValue(nonZeroVal1),
			ExpectedArg: argTest{
				IsOption:        true,
				IsProvided:      false,
				IsValueOptional: true,
				IsValueProvided: false,
				Value:           toValue(zeroVal),
			},
		},
		{
			Args:            []string{"--provided-{{.Name}}-non-zero-optional-1-value={{.NonZeroVal1}}"},
			Name:            "provided-{{.Name}}-non-zero-optional-1-value",
			ShortName:       "",
			Default:         zeroVal,
			OptionalDefault: toValue(nonZeroVal1),
			ExpectedArg: argTest{
				IsOption:        true,
				IsProvided:      true,
				IsValueOptional: true,
				IsValueProvided: false,
				Value:           toValue(nonZeroVal1),
			},
		},
		{
			Args:            []string{"--provided-{{.Name}}-non-zero-optional-2-value={{.NonZeroVal2}}"},
			Name:            "provided-{{.Name}}-non-zero-optional-2-value",
			ShortName:       "",
			Default:         zeroVal,
			OptionalDefault: toValue(nonZeroVal1),
			ExpectedArg: argTest{
				IsOption:        true,
				IsProvided:      true,
				IsValueOptional: true,
				IsValueProvided: true,
				Value:           toValue(nonZeroVal2),
			},
		},
	}
	data := struct{ Name, ZeroVal, NonZeroVal1, NonZeroVal2 string }{
		Name:        name,
		ZeroVal:     toValue(zeroVal)[0],
		NonZeroVal1: toValue(nonZeroVal1)[0],
		NonZeroVal2: toValue(nonZeroVal2)[0],
	}
	replace := func(s string) string {
		t := template.Must(template.New("test").Parse(s))
		var b bytes.Buffer
		if err := t.Execute(&b, data); err != nil {
			panic(err)
		}
		return b.String()
	}
	for _, t := range newTests {
		for k, v := range t.Args {
			t.Args[k] = replace(v)
		}
		t.Name = replace(t.Name)
		t.ExpectedArg.Name = t.Name
		t.ShortName = replace(t.ShortName)
		t.ExpectedArg.ShortName = t.ShortName
		optTests = append(optTests, t)
	}
	addSliceTests(name, zeroVal, nonZeroVal1)
}

func addSliceTests(name string, zeroVal, nonZeroVal interface{}) {
	newTests := []optTest{
		{
			Args:            nil,
			Name:            "unprovided-{{.Name}}-slice",
			ShortName:       "",
			Default:         toSlice(zeroVal, zeroVal),
			OptionalDefault: nil,
			ExpectedArg: argTest{
				IsOption:        true,
				IsProvided:      false,
				IsValueOptional: false,
				IsValueProvided: false,
				Value:           toValue(zeroVal, zeroVal),
			},
		},
		{
			Args:            nil,
			Name:            "unprovided-{{.Name}}-slice-optional",
			ShortName:       "",
			Default:         toSlice(zeroVal, zeroVal),
			OptionalDefault: toValue(zeroVal, nonZeroVal),
			ExpectedArg: argTest{
				IsOption:        true,
				IsProvided:      false,
				IsValueOptional: true,
				IsValueProvided: false,
				Value:           toValue(zeroVal, zeroVal),
			},
		},
		{
			Args:            []string{"--provided-{{.Name}}-slice-1-non-zero={{.NonZeroVal}}"},
			Name:            "provided-{{.Name}}-slice-1-non-zero",
			ShortName:       "",
			Default:         toSlice(zeroVal, zeroVal),
			OptionalDefault: nil,
			ExpectedArg: argTest{
				IsOption:        true,
				IsProvided:      true,
				IsValueOptional: false,
				IsValueProvided: true,
				Value:           toValue(nonZeroVal),
			},
		},
		{
			Args:            []string{"--provided-{{.Name}}-slice-optional-1-non-zero={{.NonZeroVal}}"},
			Name:            "provided-{{.Name}}-slice-optional-1-non-zero",
			ShortName:       "",
			Default:         toSlice(zeroVal, zeroVal),
			OptionalDefault: toValue(zeroVal, nonZeroVal),
			ExpectedArg: argTest{
				IsOption:        true,
				IsProvided:      true,
				IsValueOptional: true,
				IsValueProvided: true,
				Value:           toValue(nonZeroVal),
			},
		},
		{
			Args:            []string{"--provided-{{.Name}}-slice-2-non-zero={{.NonZeroVal}}", "--provided-{{.Name}}-slice-2-non-zero={{.NonZeroVal}}"},
			Name:            "provided-{{.Name}}-slice-2-non-zero",
			ShortName:       "",
			Default:         toSlice(zeroVal, zeroVal),
			OptionalDefault: nil,
			ExpectedArg: argTest{
				IsOption:        true,
				IsProvided:      true,
				IsValueOptional: false,
				IsValueProvided: true,
				Value:           toValue(nonZeroVal, nonZeroVal),
			},
		},
		{
			Args:            []string{"--provided-{{.Name}}-slice-optional-2-non-zero={{.NonZeroVal}}", "--provided-{{.Name}}-slice-optional-2-non-zero={{.NonZeroVal}}"},
			Name:            "provided-{{.Name}}-slice-optional-2-non-zero",
			ShortName:       "",
			Default:         toSlice(zeroVal, zeroVal),
			OptionalDefault: toValue(zeroVal, nonZeroVal),
			ExpectedArg: argTest{
				IsOption:        true,
				IsProvided:      true,
				IsValueOptional: true,
				IsValueProvided: true,
				Value:           toValue(nonZeroVal, nonZeroVal),
			},
		},
		{
			Args:            []string{"--provided-{{.Name}}-slice-1-non-zero-1-zero={{.NonZeroVal}}", "--provided-{{.Name}}-slice-1-non-zero-1-zero={{.ZeroVal}}"},
			Name:            "provided-{{.Name}}-slice-1-non-zero-1-zero",
			ShortName:       "",
			Default:         toSlice(zeroVal, zeroVal),
			OptionalDefault: nil,
			ExpectedArg: argTest{
				IsOption:        true,
				IsProvided:      true,
				IsValueOptional: false,
				IsValueProvided: true,
				Value:           toValue(nonZeroVal, zeroVal),
			},
		},
		{
			Args:            []string{"--provided-{{.Name}}-slice-optional-1-non-zero-1-zero={{.NonZeroVal}}", "--provided-{{.Name}}-slice-optional-1-non-zero-1-zero={{.ZeroVal}}"},
			Name:            "provided-{{.Name}}-slice-optional-1-non-zero-1-zero",
			ShortName:       "",
			Default:         toSlice(zeroVal, zeroVal),
			OptionalDefault: toValue(zeroVal, nonZeroVal),
			ExpectedArg: argTest{
				IsOption:        true,
				IsProvided:      true,
				IsValueOptional: true,
				IsValueProvided: true,
				Value:           toValue(nonZeroVal, zeroVal),
			},
		},
		{
			Args:            []string{"--provided-{{.Name}}-slice-optional"},
			Name:            "provided-{{.Name}}-slice-optional",
			ShortName:       "",
			Default:         toSlice(zeroVal, zeroVal),
			OptionalDefault: toValue(zeroVal, nonZeroVal),
			ExpectedArg: argTest{
				IsOption:        true,
				IsProvided:      true,
				IsValueOptional: true,
				IsValueProvided: false,
				Value:           toValue(zeroVal, nonZeroVal),
			},
		},
	}
	data := struct{ Name, ZeroVal, NonZeroVal string }{
		Name:       name,
		ZeroVal:    toValue(zeroVal)[0],
		NonZeroVal: toValue(nonZeroVal)[0],
	}
	replace := func(s string) string {
		t := template.Must(template.New("test").Parse(s))
		var b bytes.Buffer
		if err := t.Execute(&b, data); err != nil {
			panic(err)
		}
		return b.String()
	}
	for _, t := range newTests {
		t.IsSlice = true
		for k, v := range t.Args {
			t.Args[k] = replace(v)
		}
		t.Name = replace(t.Name)
		t.ExpectedArg.Name = t.Name
		t.ShortName = replace(t.ShortName)
		t.ExpectedArg.ShortName = t.ShortName
		optTests = append(optTests, t)
	}
}

var posTests = []posTest{
	{
		Name:         "no-args",
		Args:         nil,
		ExpectedArgs: nil,
	},
	{
		Name: "1-arg",
		Args: []string{"foo"},
		ExpectedArgs: []argTest{
			{
				Value: []string{"foo"},
			},
		},
	},
	{
		Name: "2-args",
		Args: []string{"foo", "bar"},
		ExpectedArgs: []argTest{
			{
				Value: []string{"foo"},
			},
			{
				Value: []string{"bar"},
			},
		},
	},
	{
		Name: "2-args-terminated",
		Args: []string{"foo", "bar", "--", "baz"},
		ExpectedArgs: []argTest{
			{
				Value: []string{"foo"},
			},
			{
				Value: []string{"bar"},
			},
			{
				Value: []string{"baz"},
			},
		},
	},
	{
		Name: "2-args-terminated-pseudo-option",
		Args: []string{"foo", "bar", "--", "--option=value"},
		ExpectedArgs: []argTest{
			{
				Value: []string{"foo"},
			},
			{
				Value: []string{"bar"},
			},
			{
				Value: []string{"--option=value"},
			},
		},
	},
}

func initPosTests() {
	for i, pt := range posTests {
		for j, arg := range pt.ExpectedArgs {
			arg.IsProvided = true
			arg.IsValueOptional = false
			arg.IsValueProvided = true
			posTests[i].ExpectedArgs[j] = arg
		}
	}
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
	addScalarTests("byte", []byte{}, []byte{1}, []byte{1, 2}) // Hex
	addScalarTests("duration", 0*time.Second, 1*time.Minute, 2*time.Hour)
	addScalarTests("ip", net.IPv4zero, net.IPv4(8, 8, 8, 8), net.IPv4(1, 1, 1, 1))
	mustParseCIDR := func(s string) net.IPNet {
		_, n, err := net.ParseCIDR(s)
		if err != nil {
			panic(err)
		}
		return *n
	}
	addScalarTests("ipnet", net.IPNet{
		IP:   net.IPv4zero,
		Mask: net.IPv4Mask(0, 0, 0, 0),
	}, mustParseCIDR("8.8.8.8/24"), mustParseCIDR("1.1.1.1/8"))
	initPosTests()
}
