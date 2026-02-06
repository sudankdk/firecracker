Purpose	Method	Endpoint
Machine config	PUT	/machine-config
Kernel	PUT	/boot-source
Rootfs	PUT	/drives/rootfs
Start VM	PUT	/actions



# start firecracker (yo terminal 1 ma hunu parxa)

# Option 1: Use the helper script (recommended)
bash start-firecracker.sh

# Option 2: Manual setup
# Clean start - kill any existing processes
sudo pkill -9 firecracker
sudo rm -f /tmp/firecracker.sock

# Start firecracker
sudo ./firecracker --api-sock /tmp/firecracker.sock

# In another terminal, fix socket permissions (do this immediately):
sudo chmod 666 /tmp/firecracker.sock

# THEN run your Go program to configure and start the VM


# Configure Maching (terminal 2)
curl --unix-socket /tmp/firecracker.sock -X PUT \
  http://localhost/machine-config \
  -H "Content-Type: application/json" \
  -d '{
    "vcpu_count": 1,
    "mem_size_mib": 512,
    "smt": false
  }'


# Configure Kernel 
curl --unix-socket /tmp/firecracker.sock -X PUT \
  http://localhost/boot-source \
  -H "Content-Type: application/json" \
  -d '{
    "kernel_image_path": "/mnt/d/firecracker/hello-vmlinux.bin",
    "boot_args": "console=ttyS0 reboot=k panic=1 pci=off"
  }'


# Configure rootfs

curl --unix-socket /tmp/firecracker.sock -X PUT \
  http://localhost/drives/rootfs \
  -H "Content-Type: application/json" \
  -d '{
    "drive_id": "rootfs",
    "path_on_host": "/mnt/d/firecracker/hello-rootfs.ext4",
    "is_root_device": true,
    "is_read_only": false
  }'

  # Start the VM 
  curl --unix-socket /tmp/firecracker.sock -X PUT \
  http://localhost/actions \
  -H "Content-Type: application/json" \
  -d '{ "action_type": "InstanceStart" }'


  wsl
sudo ./firecracker --api-sock /tmp/firecracker.sock