package cmdbuilder

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type marshaler interface {
	MarshalFlag() (string, error)
}

var (
	durationType  = reflect.TypeOf((*time.Duration)(nil)).Elem()
	marshalerType = reflect.TypeOf((*marshaler)(nil)).Elem()
)

// flagsArg represents https://github.com/jessevdk/go-flags flag
type flagsArg struct {
	isOption bool
	name     string
	tags     structTags
	typ      reflect.Type
	value    reflect.Value
}

// ArgsFromFlagsStruct converts https://github.com/jessevdk/go-flags flags from provided struct to Arg slice
func ArgsFromFlagsStruct(v interface{}) (args []Arg, err error) {
	val := reflect.ValueOf(v)
	if !val.IsValid() {
		err = errors.New("expected value, got nil")
		return
	}
	val = reflect.Indirect(val)
	if val.Kind() != reflect.Struct {
		err = errors.Errorf("expected struct, got %s", val.Kind())
		return
	}
	err = extractFlagsStructArgs(val, &args)
	return
}

func (a flagsArg) IsOption() bool { return a.isOption }

func (a flagsArg) IsProvided() bool {
	def := a.tags.All("default")
	if def == nil {
		def = a.valueToSlice(reflect.Zero(a.typ))
	}
	return !reflect.DeepEqual(a.Value(), def)
}

func (a flagsArg) IsValueOptional() bool {
	return a.isOption && (a.tags.IsTrue("optional") || a.isBoolValue())
}

func (a flagsArg) IsValueProvided() bool {
	if !a.IsProvided() {
		return false
	}
	if !a.IsValueOptional() {
		return true
	}
	return !reflect.DeepEqual(a.Value(), a.tags.All("optional-value")) && (!a.isBoolValue() || !a.isTrueValue())
}

func (a flagsArg) Name() string {
	if !a.isOption {
		return a.name
	}
	return a.tags.First("long")
}

func (a flagsArg) ShortName() string {
	if !a.isOption {
		return ""
	}
	return a.tags.First("short")
}

func (a flagsArg) Value() []string { return a.valueToSlice(a.value) }

func (a flagsArg) isBoolValue() bool {
	t := a.indirectType(a.typ)
	for {
		switch t.Kind() {
		case reflect.Array, reflect.Slice:
			return a.indirectType(t.Elem()).Kind() == reflect.Bool
		default:
			return t.Kind() == reflect.Bool
		}
	}
}

func (a flagsArg) indirect(v reflect.Value) reflect.Value {
	if !v.IsValid() {
		return reflect.Value{}
	}
	for {
		switch v.Type().Kind() {
		case reflect.Interface, reflect.Ptr:
			if v.IsNil() {
				return reflect.Value{}
			}
			v = v.Elem()
		default:
			return v
		}
	}
}

func (a flagsArg) indirectType(t reflect.Type) reflect.Type {
	for {
		switch t.Kind() {
		case reflect.Ptr:
			t = t.Elem()
		default:
			return t
		}
	}
}

func (a flagsArg) isTrueValue() bool {
	for _, v := range a.Value() {
		if v != "true" {
			return false
		}
	}
	return true
}

func (a flagsArg) valueToSlice(v reflect.Value) []string {
	v = a.indirect(v)
	if !v.IsValid() {
		return nil
	}
	var list []string
	switch v.Type().Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			list = append(list, a.valueToString(v.Index(i)))
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			list = append(list, a.valueToString(k)+":"+a.valueToString(v.MapIndex(k)))
		}
	default:
		if s := a.valueToString(v); s != "" {
			list = append(list, s)
		}
	}
	return list
}

func (a flagsArg) valueToString(v reflect.Value) string {
	v = a.indirect(v)
	if !v.IsValid() {
		return ""
	}
	typ := v.Type()
	if typ == marshalerType {
		s, err := v.Interface().(marshaler).MarshalFlag()
		if err != nil {
			panic(errors.Wrapf(err, "failed to marshal value %q", v))
		}
		return s
	}
	if typ == durationType {
		return v.Interface().(fmt.Stringer).String()
	}
	switch typ.Kind() {
	case reflect.Bool:
		if v.Bool() {
			return "true"
		}
		return "false"
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'g', -1, typ.Bits())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.String:
		return v.String()
	}
	return ""
}

func extractFlagsStructArgs(v reflect.Value, args *[]Arg) error {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.PkgPath != "" && !sf.Anonymous {
			continue
		}
		tags, err := parseStructTags(sf.Tag)
		if err != nil {
			return err
		}
		if tags.First("no-flag") != "" {
			continue
		}
		fv := v.Field(i)
		if !fv.IsValid() {
			continue
		}
		fv = reflect.Indirect(fv)
		if fv.Kind() == reflect.Struct {
			if tags.IsTrue("positional-args") {
				for j := 0; j < fv.NumField(); j++ {
					fsf := sf.Type.Field(j)
					fTags, err := parseStructTags(fsf.Tag)
					if err != nil {
						return err
					}
					*args = append(*args, &flagsArg{
						isOption: false,
						name:     stringOrDefault(fTags.First("positional-arg-name"), fsf.Name),
						typ:      sf.Type,
						value:    fv.Field(j),
					})
				}
			} else {
				if err = extractFlagsStructArgs(fv, args); err != nil {
					return err
				}
			}
			continue
		}
		if tags.First("short") == "" && tags.First("long") == "" {
			continue
		}
		*args = append(*args, &flagsArg{
			isOption: true,
			tags:     tags,
			typ:      sf.Type,
			value:    fv,
		})
	}
	return nil
}

type structTags map[string][]string

func parseStructTags(tag reflect.StructTag) (structTags, error) {
	origTag := tag
	tags := map[string][]string{}
	// Credits to reflect.StructTag.Lookup
	for tag != "" {
		// Skip leading space
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// Scan to colon
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			return nil, errors.Errorf("failed to parse struct tag %q", origTag)
		}
		name := string(tag[:i])
		tag = tag[i+1:]

		// Scan quoted string to find value
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			return nil, errors.Errorf("failed to parse struct tag %q", origTag)
		}
		val := string(tag[:i+1])
		tag = tag[i+1:]
		val, err := strconv.Unquote(val)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse struct tag %q", origTag)
		}
		tags[name] = append(tags[name], val)
	}
	return tags, nil
}

func (s structTags) All(key string) []string {
	var list []string
	if val, ok := s[key]; ok {
		for _, v := range val {
			if len(val) > 1 || v != "" {
				list = append(list, v)
			}
		}
	}
	return list
}

func (s structTags) First(key string) string {
	if v, ok := s[key]; ok && len(v) > 0 {
		return v[len(v)-1]
	}
	return ""
}

func (s structTags) IsTrue(key string) bool {
	v := s.First(key)
	return v != "" && v != "false" && v != "no" && v != "0"
}
