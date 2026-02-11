package sandboxing

import (
	"os/exec"

	"github.com/sudankdk/firecracker/internal/domain"
)

func SetUpJailer(vm *domain.VM) error {
	cmd := exec.Command(
		"./jailer",
		"--id", vm.ID,
		"--exec-file", "./firecracker",
		"--uid", "1000",
		"--gid", "1000",
		"--chroot-base-dir", "/tmp/firecracker",
		"--api-sock", vm.APISock,
	)
	return cmd.Run()
}
