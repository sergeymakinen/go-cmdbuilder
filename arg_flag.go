package cmdbuilder

import "flag"

// flagArg represents Go flag
type flagArg struct {
	flag *flag.Flag
}

// ArgsFromFlagSet converts Go flag set to Arg slice
func ArgsFromFlagSet(set *flag.FlagSet) (args []Arg, err error) {
	set.VisitAll(func(f *flag.Flag) {
		args = append(args, &flagArg{flag: f})
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

func (a flagArg) IsOption() bool { return true }

func (a flagArg) IsProvided() bool { return a.flag.Value.String() != a.flag.DefValue }

func (a flagArg) IsValueOptional() bool { return a.isBoolValue() }

func (a flagArg) IsValueProvided() bool {
	if !a.IsProvided() {
		return false
	}
	if !a.IsValueOptional() {
		return true
	}
	return a.flag.Value.String() != a.flag.DefValue && (!a.isBoolValue() || a.flag.Value.String() != "true")
}

func (a flagArg) Name() string { return a.flag.Name }

func (a flagArg) ShortName() string {
	if len(a.flag.Name) == 1 {
		return a.flag.Name
	}
	return ""
}

func (a flagArg) Value() []string {
	if s := a.flag.Value.String(); s != "" {
		return []string{s}
	}
	return nil
}

func (a flagArg) isBoolValue() bool {
	in, ok := a.flag.Value.(interface {
		IsBoolFlag() bool
	})
	return ok && in.IsBoolFlag()
}
