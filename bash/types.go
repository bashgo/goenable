package bash

import "C"

const (
	EXECUTION_SUCCESS = iota
	EXECUTION_FAILURE = iota
	EXECUTION_USAGE   = iota
)

type Enable struct {
	Name     string
	LongDoc  []string
	ShortDoc string
}

/*
Used to 'dynamically' compute a value of a variable, limited to strings right now
*/
type ValueFunc func(name string) string

func CreateVar(name string, f ValueFunc) {

}
