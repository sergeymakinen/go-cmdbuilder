package cmdbuilder

import (
	"fmt"

	"github.com/jessevdk/go-flags"
)

// ExampleCommandLine demonstrates the usage of CommandLine using an example struct
// similar to the flags package example.
func ExampleCommandLine() {
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
	cmd, _ := CommandLine(opts)
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
	// Command line: /v /v /offset 5 /n Me /animal dog /p 3 /s hello /s world /ptrslice hello /ptrslice world /intmap a:1 /intmap b:5 arg1 arg2 arg3
}
