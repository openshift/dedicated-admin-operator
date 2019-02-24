package controller

import (
	"github.com/rogbas/dedicated-admin-operator/pkg/controller/rolebinding"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, rolebinding.Add)
}
