package cmdbuilder

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

// pflagArg represents https://github.com/spf13/pflag flag
type pflagArg struct {
	set  *pflag.FlagSet
	flag *pflag.Flag
}

// ArgsFromPFlagSet converts https://github.com/spf13/pflag flag set to Arg slice
func ArgsFromPFlagSet(set *pflag.FlagSet) (args []Arg, err error) {
	set.VisitAll(func(f *pflag.Flag) {
		args = append(args, &pflagArg{
			set:  set,
			flag: f,
		})
	})
	for _, v := range set.Args() {
		if v == "" {
			args = append(args, &arg{})
		} else {
			args = append(args, &arg{value: []string{v}})
		}
	}
	return
}

func (a pflagArg) IsOption() bool { return true }

func (a pflagArg) IsProvided() bool { return a.flag.Value.String() != a.flag.DefValue }

func (a pflagArg) IsValueOptional() bool {
	return a.flag.Value.Type() == "bool" || a.flag.Value.Type() == "boolSlice" || a.flag.NoOptDefVal != ""
}

func (a pflagArg) IsValueProvided() bool {
	if !a.IsProvided() {
		return false
	}
	return !a.IsValueOptional() || strings.Join(a.Value(), ",") != a.flag.NoOptDefVal
}

func (a pflagArg) Name() string { return a.flag.Name }

func (a pflagArg) ShortName() string { return a.flag.Shorthand }

func (a pflagArg) Value() []string {
	in, ok := a.flag.Value.(interface {
		GetSlice() []string
	})
	if ok {
		return in.GetSlice()
	}
	switch a.flag.Value.Type() {
	case "boolSlice":
		list, err := a.set.GetBoolSlice(a.flag.Name)
		if err != nil {
			panic(err)
		}
		return a.listToValue(list)
	case "intSlice":
		list, err := a.set.GetIntSlice(a.flag.Name)
		if err != nil {
			panic(err)
		}
		return a.listToValue(list)
	case "uintSlice":
		list, err := a.set.GetUintSlice(a.flag.Name)
		if err != nil {
			panic(err)
		}
		return a.listToValue(list)
	case "stringSlice":
		list, err := a.set.GetStringSlice(a.flag.Name)
		if err != nil {
			panic(err)
		}
		return a.listToValue(list)
	case "durationSlice":
		list, err := a.set.GetDurationSlice(a.flag.Name)
		if err != nil {
			panic(err)
		}
		return a.listToValue(list)
	case "ipSlice":
		list, err := a.set.GetIPSlice(a.flag.Name)
		if err != nil {
			panic(err)
		}
		return a.listToValue(list)
	}
	if s := a.flag.Value.String(); s != "" {
		return []string{s}
	}
	return nil
}

func (a pflagArg) listToValue(list interface{}) []string {
	var result []string
	switch list.(type) {
	case []bool:
		for _, v := range list.([]bool) {
			result = append(result, strconv.FormatBool(v))
		}
	case []int:
		for _, v := range list.([]int) {
			result = append(result, fmt.Sprintf("%d", v))
		}
	case []uint:
		for _, v := range list.([]uint) {
			result = append(result, fmt.Sprintf("%d", v))
		}
	case []string:
		for _, v := range list.([]string) {
			result = append(result, v)
		}
	case []time.Duration:
		for _, v := range list.([]time.Duration) {
			result = append(result, fmt.Sprintf("%s", v))
		}
	case []net.IP:
		for _, v := range list.([]net.IP) {
			result = append(result, fmt.Sprintf("%s", v))
		}
	}
	return result
}
