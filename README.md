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
import "fmt"

type Options struct {
    Verbose    []bool            `short:"v" long:"verbose"`
    AuthorInfo map[string]string `short:"a"`
    Name       string            `long:"name" optional:"true"`
}

func main() {
	opts := Options{
		Verbose:    []bool{true, true, true},
		AuthorInfo: map[string]string{"name": "Jesse", "surname": "van den Kieboom"},
		Name:       "Sergey Makinen",
	}

	cmd, _ := CommandLine(opts)
	fmt.Println(cmd)
	// Output: -vvv -a name:Jesse -a "surname:van den Kieboom" --name="Sergey Makinen"
}
```

## License

BSD 3-Clause
