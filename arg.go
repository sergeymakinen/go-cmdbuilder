package cmdbuilder

// Arg represents a command-line argument
type Arg interface {
	IsOption() bool        // Whether argument is option or positional
	IsProvided() bool      // Whether argument was specified
	IsValueOptional() bool // Whether argument value is optional and can be omitted
	IsValueProvided() bool // Whether argument value was specified
	Name() string          // Argument long name
	ShortName() string     // Argument one char shorthand
	Value() []string       // Argument value
}

// arg represents a non-option argument
type arg struct {
	value []string
}

func (a arg) IsOption() bool        { return false }
func (a arg) IsProvided() bool      { return true }
func (a arg) IsValueOptional() bool { return false }
func (a arg) IsValueProvided() bool { return true }
func (a arg) Name() string          { return "" }
func (a arg) ShortName() string     { return "" }
func (a arg) Value() []string       { return a.value }
