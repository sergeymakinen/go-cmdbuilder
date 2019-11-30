package cmdbuilder

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/jessevdk/go-flags"
	"github.com/ompluscator/dynamic-struct"
)

func ExampleArgsFromFlagsStruct() {
	s := struct {
		Agree bool `long:"agree"`
		Age   uint `long:"age" short:"a"`
	}{}
	flags.ParseArgs(&s, []string{"--agree", "-a", "18"})
	args, _ := ArgsFromFlagsStruct(s)
	list, _ := Build(args)
	fmt.Println(strings.Join(list, " "))
	// Output: --agree -a 18
}

func TestArgsFromFlagsStructWithOptions(t *testing.T) {
	allowedErrors := regexp.MustCompile("boolean flag .+ may not have default values")
	for _, ft := range optTests {
		t.Run("Name="+ft.Name, func(t *testing.T) {
			tag := `short:"` + ft.ShortName + `" long:"` + ft.Name + `"`
			if ft.IsSlice {
				v := reflect.ValueOf(ft.Default)
				for i := 0; i < v.Len(); i++ {
					tag += ` default:"` + toValue(v.Index(i).Interface())[0] + `"`
				}
			} else {
				for _, v := range toValue(ft.Default) {
					tag += ` default:"` + v + `"`
				}
			}
			if ft.OptionalDefault != nil {
				tag += ` optional:"true"`
				for _, v := range ft.OptionalDefault {
					tag += ` optional-value:"` + v + `"`
				}
			}
			var s interface{}
			switch ft.Default.(type) {
			case bool, []bool, float32, []float32, float64, []float64, int8, []int8, uint8, []uint8, int16, []int16, uint16, []uint16, int32, []int32, uint32, []uint32, int64, []int64, uint64, []uint64, int, []int, uint, []uint, string, []string:
				if !strings.Contains(ft.Name, "-byte-") && !strings.HasSuffix(ft.Name, "-byte") {
					s = dynamicstruct.NewStruct().
						AddField("Value", ft.Default, tag).
						Build().
						New()
				} else {
					t.Skip("reflect.ValueOf(optTest.Default).Type() = []byte")
				}
			default:
				t.Skipf("reflect.ValueOf(optTest.Default).Type() = %s", reflect.ValueOf(ft.Default).Type())
			}
			if _, err := flags.ParseArgs(s, ft.Args); err != nil {
				if allowedErrors.MatchString(err.Error()) {
					t.Skipf("ParseArgs(optTest.Args) = %v", err)
				} else {
					t.Fatalf("ParseArgs(optTest.Args) = %v", err)
				}
			}
			args, err := ArgsFromFlagsStruct(s)
			if err != nil {
				t.Fatalf("ArgsFromFlagsStruct(set) = %v", err)
			}
			if len(args) != 1 {
				t.Fatalf("len(args) = %d", len(args))
			}
			ft.ExpectedArg.TestEqual(t, args[0])
		})
	}
}

func TestArgsFromFlagsStructShouldFailOnNonStructs(t *testing.T) {
	if _, err := ArgsFromFlagsStruct(nil); err == nil {
		t.Error("ArgsFromFlagsStruct(nil) = nil; want non-nil")
	}
	if _, err := ArgsFromFlagsStruct(true); err == nil {
		t.Error("ArgsFromFlagsStruct(true) = nil; want non-nil")
	}
	badStruct := struct {
		Bad string `malformed`
	}{}
	if _, err := ArgsFromFlagsStruct(badStruct); err == nil {
		t.Error("ArgsFromFlagsStruct(badStruct) = nil; want non-nil")
	}
}

func TestArgsFromFlagsStructWithPositional(t *testing.T) {
	for _, pt := range posTests {
		t.Run("Name="+pt.Name, func(t *testing.T) {
			s := struct {
				Positional struct {
					Value1, Value2, Value3 string
				} `positional-args:"yes"`
			}{}
			if _, err := flags.ParseArgs(&s, pt.Args); err != nil {
				t.Fatalf("ParseArgs(posTest.Args) = %v", err)
			}
			args, err := ArgsFromFlagsStruct(s)
			if err != nil {
				t.Fatalf("ArgsFromFlagsStruct(set) = %v", err)
			}
			if len(args) < len(pt.ExpectedArgs) {
				t.Fatalf("len(args) = %d; want %d or more", len(args), len(pt.ExpectedArgs))
			}
			for i, arg := range pt.ExpectedArgs {
				arg.TestEqual(t, args[i])
			}
		})
	}
}
