package handler

import (
	"github.com/morvencao/minicni/pkg/args"
)

type Handler interface {
	HandleAdd(cmdArgs *args.CmdArgs) error
	HandleDel(cmdArgs *args.CmdArgs) error
	HandleCheck(cmdArgs *args.CmdArgs) error
	HandleVersion(cmdArgs *args.CmdArgs) error
}
