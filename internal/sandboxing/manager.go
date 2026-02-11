package sandboxing

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sudankdk/firecracker/internal/domain"
)

type VMManager struct {
	BaseChrootDir string // e.g., "/srv/vms"
	BaseUploadDir string // e.g., "/srv/uploads"
	KernelPath    string
	RootfsPath    string
}

func (mgr *VMManager) SpawnVM(uploadFilePath string) (*domain.VM, error) {
	vm, err := CreateVMMetadata()
	if err != nil {
		return nil, err
	}
	chrootDir := filepath.Join(mgr.BaseChrootDir, vm.ID)
	if err := os.MkdirAll(chrootDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create chroot: %w", err)
	}

	inputDrive := filepath.Join(chrootDir, "input_drive.img")
	if err := copyFile(uploadFilePath, inputDrive); err != nil {
		return nil, fmt.Errorf("failed to copy upload: %w", err)
	}

	if err := CreateTAP(vm.TapName); err != nil {
		return nil, fmt.Errorf("failed to create TAP: %w", err)
	}

	cmd, err := mgr.SetUpJailer(vm)
	if err != nil {
		return nil, fmt.Errorf("failed to set up jailer: %w", err)
	}
	vm.Cmd = cmd

	if err := configureVM(vm, mgr.KernelPath, mgr.RootfsPath, inputDrive); err != nil {
		return nil, fmt.Errorf("failed to configure VM: %w", err)
	}

	// Step 7: Cleanup after VM exits
	go func() {
		cmd.Wait()
		_ = exec.Command("ip", "link", "del", vm.TapName).Run()
		os.Remove(vm.APISock)
		os.RemoveAll(chrootDir)
	}()

	return vm, nil

}

// copyFile safely copies a file
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func (mgr *VMManager) SetUpJailer(vm *domain.VM) (*exec.Cmd, error) {
	cmd := exec.Command(
		"./jailer",
		"--id", vm.ID,
		"--exec-file", "./firecracker",
		"--uid", "1000",
		"--gid", "1000",
		"--chroot-base-dir", mgr.BaseChrootDir,
		"--api-sock", vm.APISock,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start Firecracker: %w", err)
	}
	return cmd, nil
}
