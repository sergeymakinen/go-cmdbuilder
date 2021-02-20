# cmdbuilder

[![tests](https://github.com/sergeymakinen/go-cmdbuilder/workflows/tests/badge.svg)](https://github.com/sergeymakinen/go-cmdbuilder/actions?query=workflow%3Atests)
[![Go Reference](https://pkg.go.dev/badge/github.com/sergeymakinen/go-cmdbuilder.svg)](https://pkg.go.dev/github.com/sergeymakinen/go-cmdbuilder/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/sergeymakinen/go-cmdbuilder)](https://goreportcard.com/report/github.com/sergeymakinen/go-cmdbuilder)
[![codecov](https://codecov.io/gh/sergeymakinen/go-cmdbuilder/branch/main/graph/badge.svg)](https://codecov.io/gh/sergeymakinen/go-cmdbuilder)

Package cmdbuilder implements a converter from structs defining command-line arguments and their values back to a command-line.

The cmdbuilder package aims to be compatible with the flags package (see https://github.com/jessevdk/go-flags),
so it also uses structs, the reflection and struct field tags to specify command-line arguments. For example:

```go
type Options struct {
    Verbose    []bool            `short:"v" long:"verbose"`
    AuthorInfo map[string]string `short:"a"`
    Name       string            `long:"name" optional:"true"`
}
```

This specifies the `Verbose` boolean option with a short name `-v` and a long name `--verbose`,
the `AuthorInfo` map option with a short name `-a`,
and the `Name` string option with a long name `--name` and an optional value.
If the struct is initialized as the following:

```go
opts := Options{
    Verbose:    []bool{true, true, true},
    AuthorInfo: map[string]string{"name": "Jesse", "surname": "van den Kieboom"},
    Name:       "Sergey Makinen",
}
```

Then the `CommandLine` function will produce the following string:

```
-vvv -a name:Jesse -a "surname:van den Kieboom" --name="Sergey Makinen"
```

Any type that implements the `Marshaler` interface may fully customize its value output.

## Installation

Use go get:

```bash
go get github.com/sergeymakinen/go-cmdbuilder/v2
```

Then import the package into your own code:

```go
import "github.com/sergeymakinen/go-cmdbuilder/v2"
```


## Example

```go
package main

import (
	"fmt"

	"github.com/jessevdk/go-flags"
	"github.com/sergeymakinen/go-cmdbuilder/v2"
)

type Options struct {
	// Slice of bool will append 'true' each time the option
	// is encountered (can be set multiple times, like -vvv)
	Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`

	// Example of automatic marshalling to desired type (uint)
	Offset uint `long:"offset" description:"Offset"`

	// Example of a required flag
	Name string `short:"n" long:"name" description:"A name" required:"true"`

	// Example of a flag restricted to a pre-defined set of strings
	Animal string `long:"animal" choice:"cat" choice:"dog"`

	// Example of a value name
	File string `short:"f" long:"file" description:"A file" value-name:"FILE"`

	// Example of a pointer
	Ptr *int `short:"p" description:"A pointer to an integer"`

	// Example of a slice of strings
	StringSlice []string `short:"s" description:"A slice of strings"`

	// Example of a slice of pointers
	PtrSlice []*string `long:"ptrslice" description:"A slice of pointers to string"`

	// Example of a map
	IntMap map[string]int `long:"intmap" description:"A map from string to int"`

	// Example of positional arguments
	Positional struct {
		Rest []string
	} `positional-args:"yes"`
}

// main demonstrates the usage of CommandLine using an example struct
// similar to the flags package example.
func main() {
	var opts Options
	args := []string{
		"-vv",
		"--offset=5",
		"-n", "Me",
		"--animal", "dog",
		"-p", "3",
		"-s", "hello",
		"-s", "world",
		"--ptrslice", "hello",
		"--ptrslice", "world",
		"--intmap", "a:1",
		"--intmap", "b:5",
		"arg1",
		"arg2",
		"arg3",
	}
	_, err := flags.ParseArgs(&opts, args)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Verbose: %+v\nOffset: %+v\nName: %+v\nAnimal: %+v\nPtr: %+v\nStringSlice: %+v\nPtrSlice: [%v %v]\nIntMap: %+v\nPositional: %+v\n\n", opts.Verbose, opts.Offset, opts.Name, opts.Animal, *opts.Ptr, opts.StringSlice, *opts.PtrSlice[0], *opts.PtrSlice[1], opts.IntMap, opts.Positional)
	cmd, _ := cmdbuilder.CommandLine(opts)
	fmt.Printf("Command line: %s\n", cmd)
	// Output:
	// Verbose: [true true]
	// Offset: 5
	// Name: Me
	// Animal: dog
	// Ptr: 3
	// StringSlice: [hello world]
	// PtrSlice: [hello world]
	// IntMap: map[a:1 b:5]
	// Positional: {Rest:[arg1 arg2 arg3]}
	//
	// Command line: -vv --offset 5 -n Me --animal dog -p 3 -s hello -s world --ptrslice hello --ptrslice world --intmap a:1 --intmap b:5 arg1 arg2 arg3
}
```

## License

BSD 3-Clause
