package cmdbuilder

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"time"

	"github.com/fatih/structtag"
	"github.com/pkg/errors"
)

// Marshaler is the interface implemented by types that can marshal themselves
// to a string representation of the flag.
type Marshaler interface {
	MarshalFlag() (string, error)
}

var (
	durationType  = reflect.TypeOf((*time.Duration)(nil)).Elem()
	marshalerType = reflect.TypeOf((*Marshaler)(nil)).Elem()
)

// arg wraps struct field flag.
type arg struct {
	isOption bool
	st       reflect.Type
	sf       reflect.StructField
	tags     *structTags
	value    reflect.Value
}

func (a arg) Struct() reflect.Type { return a.st }

func (a arg) Field() reflect.StructField { return a.sf }

func (a arg) IsOption() bool { return a.isOption }

func (a arg) IsProvided() bool {
	def := a.tags.All("default")
	if def == nil {
		def = valueSlice(reflect.Zero(a.sf.Type))
	}
	return !reflect.DeepEqual(a.Value(), def)
}

func (a arg) IsValueOptional() bool {
	return a.isOption && (a.tags.IsTrue("optional") || a.isBoolean())
}

func (a arg) IsValueProvided() bool {
	if !a.IsProvided() {
		return false
	}
	if !a.IsValueOptional() {
		return true
	}
	return !reflect.DeepEqual(a.Value(), a.tags.All("optional-value")) && (!a.isBoolean() || !a.isTrueValue())
}

func (a arg) Name() string {
	return a.tags.First("long")
}

func (a arg) ShortName() string {
	return a.tags.First("short")
}

func (a arg) Value() []string { return valueSlice(a.value) }

func (a arg) isBoolean() bool {
	t := indirectType(a.sf.Type)
	for {
		switch t.Kind() {
		case reflect.Array, reflect.Slice:
			return indirectType(t.Elem()).Kind() == reflect.Bool
		default:
			return t.Kind() == reflect.Bool
		}
	}
}

func (a arg) isTrueValue() bool {
	for _, v := range a.Value() {
		if v != "true" {
			return false
		}
	}
	return true
}

func parse(v interface{}) (args []arg, err error) {
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
	err = parseStruct(val, &args)
	return
}

func parseStruct(v reflect.Value, args *[]arg) error {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.PkgPath != "" && !sf.Anonymous {
			continue
		}
		tags, err := parseStructTags(sf.Tag)
		if err != nil {
			return &FieldError{
				Struct: t,
				Field:  sf.Name,
				Type:   sf.Type,
				Msg:    err.Error(),
			}
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
					*args = append(*args, arg{
						isOption: false,
						st:       v.Type(),
						sf:       sf,
						value:    fv.Field(j),
					})
				}
			} else {
				if err = parseStruct(fv, args); err != nil {
					return &FieldError{
						Struct: t,
						Field:  sf.Name,
						Type:   sf.Type,
						Msg:    err.Error(),
					}
				}
			}
			continue
		}
		if tags.First("short") == "" && tags.First("long") == "" {
			continue
		}
		*args = append(*args, arg{
			isOption: true,
			st:       v.Type(),
			sf:       sf,
			tags:     tags,
			value:    fv,
		})
	}
	return nil
}

func indirect(v reflect.Value) reflect.Value {
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

func indirectType(t reflect.Type) reflect.Type {
	for {
		switch t.Kind() {
		case reflect.Ptr:
			t = t.Elem()
		default:
			return t
		}
	}
}

func valueSlice(v reflect.Value) []string {
	v = indirect(v)
	if !v.IsValid() {
		return nil
	}
	var list []string
	switch v.Type().Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			list = append(list, valueString(v.Index(i)))
		}
	case reflect.Map:
		keys := v.MapKeys()
		sort.Slice(keys, func(i, j int) bool {
			return valueString(keys[i]) < valueString(keys[j])
		})
		for _, k := range keys {
			list = append(list, valueString(k)+":"+valueString(v.MapIndex(k)))
		}
	default:
		if s := valueString(v); s != "" {
			list = append(list, s)
		}
	}
	return list
}

func valueString(v reflect.Value) string {
	v = indirect(v)
	if !v.IsValid() {
		return ""
	}
	typ := v.Type()
	if v.CanInterface() && typ.Implements(marshalerType) {
		s, err := v.Interface().(Marshaler).MarshalFlag()
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

type structTags struct {
	tags structtag.Tags
}

func (t *structTags) All(key string) []string {
	if t == nil {
		return nil
	}
	var list []string
	for _, tag := range t.tags.Tags() {
		if tag.Key == key {
			list = append(list, tag.Name)
		}
	}
	if len(list) == 1 && list[0] == "" {
		return nil
	}
	return list
}

func (t *structTags) First(key string) string {
	if t == nil {
		return ""
	}
	if tag, err := t.tags.Get(key); err == nil {
		return tag.Name
	}
	return ""
}

func (t *structTags) IsTrue(key string) bool {
	v := t.First(key)
	return v != "" && v != "false" && v != "no" && v != "0"
}

func parseStructTags(tag reflect.StructTag) (*structTags, error) {
	tags, err := structtag.Parse(string(tag))
	if err != nil {
		return nil, err
	}
	return &structTags{tags: *tags}, nil
}
