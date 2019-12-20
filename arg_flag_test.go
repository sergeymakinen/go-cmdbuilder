package cmdbuilder

import (
	"flag"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

func ExampleArgsFromFlagSet() {
	set := flag.NewFlagSet("options", flag.ContinueOnError)
	set.Bool("agree", false, "")
	set.Uint("age", 0, "")
	set.Parse([]string{"-agree=true", "-age", "18"})
	args, _ := ArgsFromFlagSet(set)
	list, _ := Build(args)
	fmt.Println(strings.Join(list, " "))
	// Output: --age 18 --agree
}

func TestArgsFromFlagSetWithOptions(t *testing.T) {
	for _, ot := range optTests {
		t.Run("Name="+ot.Name, func(t *testing.T) {
			set := flag.NewFlagSet("test", flag.ContinueOnError)
			switch ot.Default.(type) {
			case bool:
				set.Bool(ot.Name, ot.Default.(bool), "")
			case float64:
				set.Float64(ot.Name, ot.Default.(float64), "")
			case int:
				set.Int(ot.Name, ot.Default.(int), "")
			case int64:
				set.Int64(ot.Name, ot.Default.(int64), "")
			case uint:
				set.Uint(ot.Name, ot.Default.(uint), "")
			case uint64:
				set.Uint64(ot.Name, ot.Default.(uint64), "")
			case string:
				set.String(ot.Name, ot.Default.(string), "")
			case time.Duration:
				set.Duration(ot.Name, ot.Default.(time.Duration), "")
			default:
				t.Skipf("reflect.ValueOf(optTest.Default).Type() = %s", reflect.ValueOf(ot.Default).Type())
			}
			if ot.OptionalDefault != nil {
				t.Skipf("optTest.OptionalDefault = %v; want nil", ot.OptionalDefault)
			}
			if err := set.Parse(ot.Args); err != nil {
				t.Fatalf("FlagSet.Parse(optTest.Args) = %v; want nil", err)
			}
			args, err := ArgsFromFlagSet(set)
			if err != nil {
				t.Fatalf("ArgsFromFlagSet(set) = %v; want nil", err)
			}
			if len(args) != 1 {
				t.Fatalf("len(args) = %d; want 1", len(args))
			}
			ot.ExpectedArg.TestEqual(t, args[0])
		})
	}
}

func TestArgsFromFlagSetWithPositional(t *testing.T) {
	for _, pt := range posTests {
		t.Run("Name="+pt.Name, func(t *testing.T) {
			if strings.Contains(pt.Name, "terminated") {
				t.Skipf("posTest.Name = %s", pt.Name)
			}
			set := flag.NewFlagSet("test", flag.ContinueOnError)
			if err := set.Parse(pt.Args); err != nil {
				t.Fatalf("FlagSet.Parse(posTest.Args) = %v", err)
			}
			args, err := ArgsFromFlagSet(set)
			if err != nil {
				t.Fatalf("ArgsFromFlagSet(set) = %v; want nil", err)
			}
			if len(args) != len(pt.ExpectedArgs) {
				t.Fatalf("len(args) = %d; want %d", len(args), len(pt.ExpectedArgs))
			}
			for i, arg := range pt.ExpectedArgs {
				arg.TestEqual(t, args[i])
			}
		})
	}
}
