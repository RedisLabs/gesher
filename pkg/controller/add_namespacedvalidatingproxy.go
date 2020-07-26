package controller

import (
	"github.com/redislabs/gesher/pkg/controller/namespacedvalidatingproxy"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, namespacedvalidatingproxy.Add)
}
