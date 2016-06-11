# GoEnable #

Us bash shell function 'enable' to execute loadable builtin commands
implemented using Go.

## Requirements ##

- Go 1.5.X or later
- Bash 4.3.X or later

## Installation ##

```
go get github.com/bashgo/goenable
go install github.com/bashgo/goenable
```

## Implement Bash builtin functions with Go ##

```
//go:generate goenable -input $GOFILE

package main

import "C"

import (
"fmt"

"github.com/bashgo/goenable/bash"

)

//export mybuiltin
func mybuiltin(args []string) C.int {
for _, a := range args {
    fmt.Println(a)
  }
  return bash.EXECUTION_SUCCESS
}

//export gohello
func gohello(args []string) C.int {
  for _, a := range args {
    fmt.Println("HELLO,", a)
  }
  return bash.EXECUTION_SUCCESS
}

//export gotrue
func gotrue(args []string) C.int {
  return bash.EXECUTION_SUCCESS
}

//export gofalse
func gofalse(args []string) C.int {
  fmt.Println("GOFALSE:", bash.EXECUTION_FAILURE)
  return bash.EXECUTION_FAILURE
}

/*
this var section is parsed by the goenable command, the contents is
very strict, I.E. the parser is not very flexible at all
The fields in the bash.Enable type are:

name of the command, this must match the actual function name
array of string, displayed with bash help command
short doc string, a one liner
*/

var (
  cmd1 = bash.Enable{
    "mybuiltin",
    []string{"doc line 1", "doc line 2"},
    "mybuilt in just prints out all parameters, on on each line",
  }

  cmd2 = bash.Enable{
    "gotrue",
    []string{"like /bin/true."},
    "gotrue exists with success status",
  }

  cmd3 = bash.Enable{
    "gofalse",
    []string{"like /bin/false."},
    "gofalse exists if failure status",
  }
)

// main function is need to compile a c-shared library
func main() {
}
```


## Compile the shared library ##

```
go build -v -buildmode=c-shared
```

## Use it, with bash enable command ##

```
enable -f <fullpath to the shared library file> <name of implemented function>
```

## Example usage ##

```
bash -c 'enable -f ./goenableexample gotrue gofalse;gotrue || echo
TRUE FAILSED THIS SHOULD NOT BE DISPLAYED; gofalse || echo FALSE WAS
FALSE THIS SHOULD BE DISPLAYED'
```
