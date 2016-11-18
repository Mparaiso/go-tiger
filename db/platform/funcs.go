package platform

import (
	"github.com/Mparaiso/go-tiger/funcs"
)

var (
	mapStringsToStrings func([]string, func(string) string) []string
	_                   = funcs.Must(funcs.MakeMap(&mapStringsToStrings))
)
