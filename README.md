# cmdbuilder

[![Travis-CI](https://travis-ci.com/sergeymakinen/go-cmdbuilder.svg)](https://travis-ci.com/sergeymakinen/go-cmdbuilder) [![AppVeyor](https://ci.appveyor.com/api/projects/status/p2gum60gw4t6alji/branch/master?svg=true)](https://ci.appveyor.com/project/sergeymakinen/go-cmdbuilder/branch/master) [![GoDoc](https://godoc.org/github.com/sergeymakinen/go-cmdbuilder?status.svg)](http://godoc.org/github.com/sergeymakinen/go-cmdbuilder) [![Report card](https://goreportcard.com/badge/github.com/sergeymakinen/go-cmdbuilder)](https://goreportcard.com/report/github.com/sergeymakinen/go-cmdbuilder)

Package cmdbuilder provides an options-to-command-line arguments converter.

It's used to convert different flag sets and struct tag based flag options back to a command-line.

For now, it supports the following options:
- native Go FlagSet
- https://github.com/spf13/pflag FlagSet
- https://github.com/jessevdk/go-flags like struct tag mapped struct

## Installation

Use go get:

```bash
go get github.com/sergeymakinen/go-cmdbuilder
```

Then import the package into your own code:

```go
import "github.com/sergeymakinen/go-cmdbuilder"
```


## Example

```go
s := struct {
    Agree bool `long:"agree"`
    Age   uint `long:"age" short:"a"`
}{}
flags.ParseArgs(&s, []string{"--agree", "-a", "18"})
args, _ := ArgsFromFlagsStruct(s)
list, _ := NewBuilder(args).Build()
fmt.Println(strings.Join(list, " "))
// Output: --agree -a 18
```

## License

BSD 3-Clause
