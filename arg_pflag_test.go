package cmdbuilder

import (
	"fmt"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/spf13/pflag"
)

func ExampleArgsFromPFlagSet() {
	set := pflag.NewFlagSet("options", pflag.ContinueOnError)
	set.Bool("agree", false, "")
	set.UintP("age", "a", 0, "")
	set.Parse([]string{"--agree=true", "-a", "18"})
	args, _ := ArgsFromPFlagSet(set)
	list, _ := Build(args)
	fmt.Println(strings.Join(list, " "))
	// Output: -a 18 --agree
}

func TestArgsFromPFlagSetWithOptions(t *testing.T) {
	for _, ft := range optTests {
		t.Run("Name="+ft.Name, func(t *testing.T) {
			set := pflag.NewFlagSet("test", pflag.ContinueOnError)
			switch ft.Default.(type) {
			case bool:
				set.Bool(ft.Name, ft.Default.(bool), "")
			case []bool:
				set.BoolSlice(ft.Name, ft.Default.([]bool), "")
			case float32:
				set.Float32(ft.Name, ft.Default.(float32), "")
			case float64:
				set.Float64(ft.Name, ft.Default.(float64), "")
			case int8:
				set.Int8(ft.Name, ft.Default.(int8), "")
			case uint8:
				set.Uint8(ft.Name, ft.Default.(uint8), "")
			case int16:
				set.Int16(ft.Name, ft.Default.(int16), "")
			case uint16:
				set.Uint16(ft.Name, ft.Default.(uint16), "")
			case int32:
				set.Int32(ft.Name, ft.Default.(int32), "")
			case uint32:
				set.Uint32(ft.Name, ft.Default.(uint32), "")
			case int64:
				set.Int64(ft.Name, ft.Default.(int64), "")
			case uint64:
				set.Uint64(ft.Name, ft.Default.(uint64), "")
			case int:
				set.Int(ft.Name, ft.Default.(int), "")
			case []int:
				set.IntSlice(ft.Name, ft.Default.([]int), "")
			case uint:
				set.Uint(ft.Name, ft.Default.(uint), "")
			case []uint:
				set.UintSlice(ft.Name, ft.Default.([]uint), "")
			case string:
				set.String(ft.Name, ft.Default.(string), "")
			case []string:
				if !strings.HasSuffix(ft.Name, "1-zero") {
					set.StringSlice(ft.Name, ft.Default.([]string), "")
				} else {
					t.Skipf("optTest.ExpectedArg.Value = %s", ft.ExpectedArg.Value)
				}
			case []byte:
				if !strings.Contains(ft.Name, "uint8") {
					set.BytesHex(ft.Name, ft.Default.([]byte), "")
				} else {
					t.Skip("reflect.ValueOf(optTest.Default).Type() = []uint8")
				}
			case time.Duration:
				set.Duration(ft.Name, ft.Default.(time.Duration), "")
			case []time.Duration:
				set.DurationSlice(ft.Name, ft.Default.([]time.Duration), "")
			case net.IP:
				set.IP(ft.Name, ft.Default.(net.IP), "")
			case []net.IP:
				set.IPSlice(ft.Name, ft.Default.([]net.IP), "")
			case net.IPNet:
				set.IPNet(ft.Name, ft.Default.(net.IPNet), "")
			default:
				t.Skipf("reflect.ValueOf(optTest.Default).Type() = %s", reflect.ValueOf(ft.Default).Type())
			}
			if ft.OptionalDefault != nil {
				set.Lookup(ft.Name).NoOptDefVal = strings.Join(ft.OptionalDefault, ",")
			}
			if err := set.Parse(ft.Args); err != nil {
				t.Fatalf("FlagSet.Parse() = %v; want nil", err)
			}
			args, err := ArgsFromPFlagSet(set)
			if err != nil {
				t.Fatalf("ArgsFromPFlagSet() = _, %v; want nil", err)
			}
			if len(args) != 1 {
				t.Fatalf("len(args) = %d; want 1", len(args))
			}
			ft.ExpectedArg.TestEqual(t, args[0])
		})
	}
}

func TestArgsFromPFlagSetWithPositional(t *testing.T) {
	for _, pt := range posTests {
		t.Run("Name="+pt.Name, func(t *testing.T) {
			set := pflag.NewFlagSet("test", pflag.ContinueOnError)
			if err := set.Parse(pt.Args); err != nil {
				t.Fatalf("FlagSet.Parse() = %v; want nil", err)
			}
			args, err := ArgsFromPFlagSet(set)
			if err != nil {
				t.Fatalf("ArgsFromPFlagSet() = _, %v; want nil", err)
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
