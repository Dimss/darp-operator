package controller

import (
	"github.com/darp-operator/pkg/controller/darp"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, darp.Add)
}
