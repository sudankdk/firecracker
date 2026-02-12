package sandboxing

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/sudankdk/firecracker/internal/domain"
)

type VMManager struct {
	BaseChrootDir   string // e.g., "/srv/vms"
	BaseUploadDir   string // e.g., "/srv/uploads"
	KernelPath      string
	RootfsPath      string
	JailerPath      string
	FirecrackerPath string
}

func (mgr *VMManager) SpawnVM(uploadFilePath string) (*domain.VM, error) {
	vm, err := CreateVMMetadata()
	if err != nil {
		return nil, err
	}
	vmDir := filepath.Join(mgr.BaseChrootDir, vm.ID)
	if err := os.MkdirAll(vmDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create VM directory: %w", err)
	}

	inputDrive := filepath.Join(vmDir, "input_drive.img")
	if err := copyFile(uploadFilePath, inputDrive); err != nil {
		return nil, fmt.Errorf("failed to copy upload: %w", err)
	}

	kernelPath := filepath.Join(vmDir, "kernel")
	if err := copyFile(mgr.KernelPath, kernelPath); err != nil {
		return nil, fmt.Errorf("failed to copy kernel: %w", err)
	}

	rootfsPath := filepath.Join(vmDir, "rootfs.ext4")
	if err := copyFile(mgr.RootfsPath, rootfsPath); err != nil {
		return nil, fmt.Errorf("failed to copy rootfs: %w", err)
	}

	// TAP and networking disabled for manual isolation in WSL
	// if err := CreateTAP(vm.TapName); err != nil {
	// 	return nil, fmt.Errorf("failed to create TAP: %w", err)
	// }

	vm.APISock = filepath.Join(vmDir, "firecracker.socket")

	cmd, err := mgr.SetUpFirecracker(vm)
	if err != nil {
		return nil, fmt.Errorf("failed to set up Firecracker: %w", err)
	}
	vm.Cmd = cmd
	if err := waitForSocket(vm.APISock, 5*time.Second); err != nil {
		return nil, err
	}

	if err := configureVM(vm, kernelPath, rootfsPath, inputDrive); err != nil {
		return nil, fmt.Errorf("failed to configure VM: %w", err)
	}

	// Cleanup after VM exits
	go func() {
		cmd.Wait()
		os.RemoveAll(vmDir)
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

func waitForSocket(socketPath string, timeout time.Duration) error {
	start := time.Now()
	for {
		if _, err := os.Stat(socketPath); err == nil {
			return nil
		}
		if time.Since(start) > timeout {
			return fmt.Errorf("socket did not appear within %v", timeout)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (mgr *VMManager) SetUpFirecracker(vm *domain.VM) (*exec.Cmd, error) {
	// Jailer is not supported in WSL, so running Firecracker directly with manual isolation
	firecrackerPath := mgr.FirecrackerPath
	if firecrackerPath == "" {
		firecrackerPath = "./firecracker"
	}

	cmd := exec.Command(firecrackerPath, "--api-sock", vm.APISock)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start Firecracker: %w", err)
	}
	return cmd, nil
}
