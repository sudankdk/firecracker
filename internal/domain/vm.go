package domain

import "os/exec"

type VM struct {
	ID      string
	Cmd     *exec.Cmd
	APISock string
	TapName string
}
