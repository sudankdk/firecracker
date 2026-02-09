# Quick Start Guide

## Prerequisites Check

Before running the system, ensure you have:

1. **WSL2 with Ubuntu** (or similar Linux distribution)
2. **Go 1.25+** installed
3. **Root privileges** (required for mount operations)
4. **Firecracker binary** installed
5. **Required tools**: `fallocate`, `mkfs.ext4`, `mount`, `umount`

### Install Missing Dependencies
```bash
# Install filesystem tools
sudo apt-get update
sudo apt-get install -y util-linux e2fsprogs

# Install Firecracker (if not already installed)
# Visit: https://github.com/firecracker-microvm/firecracker/releases
wget https://github.com/firecracker-microvm/firecracker/releases/download/v1.4.0/firecracker-v1.4.0-x86_64.tgz
tar -xzf firecracker-v1.4.0-x86_64.tgz
sudo mv release-v1.4.0-x86_64/firecracker-v1.4.0-x86_64 /usr/local/bin/firecracker
sudo chmod +x /usr/local/bin/firecracker
```

---

## Required Files

You need these Firecracker images in `/mnt/d/firecracker/`:

1. **Kernel**: `hello-vmlinux.bin`
2. **Root filesystem**: `hello-rootfs.ext4`

### Download Sample Images (for testing)
```bash
cd /mnt/d/firecracker

# Download Alpine Linux kernel
wget https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/kernels/vmlinux.bin -O hello-vmlinux.bin

# Download Alpine rootfs
wget https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/rootfs/alpine.ext4 -O hello-rootfs.ext4
```

---

## Build and Run

### 1. Navigate to Project Directory
```bash
cd /mnt/d/firecracker
```

### 2. Install Go Dependencies
```bash
go mod download
```

### 3. Start Server (as root)
```bash
sudo go run *.go
```

Expected output:
```
2026/02/09 12:00:00 Initialized directory: /mnt/d/firecracker/uploads
2026/02/09 12:00:00 Initialized directory: /mnt/d/firecracker/disks
2026/02/09 12:00:00 Initialized directory: /mnt/d/firecracker/mnt
2026/02/09 12:00:00 HTTP server listening on port 8080...
2026/02/09 12:00:00 Endpoints:
2026/02/09 12:00:00   POST   /upload              - Upload file for scanning
2026/02/09 12:00:00   POST   /vm/scan/{jobID}     - Start VM to scan uploaded file
2026/02/09 12:00:00   GET    /jobs/{jobID}        - Get job status
2026/02/09 12:00:00   GET    /jobs                - List all jobs
2026/02/09 12:00:00   DELETE /jobs/{jobID}        - Cleanup job resources
2026/02/09 12:00:00   POST   /vm/start            - Start VM (legacy)
```

---

## Testing the System

### Test 1: Upload a File

**Create a test file:**
```bash
echo "This is a test file for malware scanning" > testfile.txt
```

**Upload the file:**
```bash
curl -F "file=@testfile.txt" http://localhost:8080/upload
```

**Expected Response:**
```json
{
  "jobID": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "hash": "5d41402abc4b2a76b9719d911017c592ae2e1edc318210e3e8e5e6f8b092c3e2",
  "status": "disk_created"
}
```

**Verify:**
- Check server logs for hash computation
- Verify disk image exists:
  ```bash
  ls -lh /mnt/d/firecracker/disks/
  # Should show: input-{jobID}.ext4 (50MB)
  ```

**Verify hash matches:**
```bash
sha256sum testfile.txt
# Should match the hash in the response
```

---

### Test 2: Check Job Status

```bash
# Replace {jobID} with the jobID from upload response
curl http://localhost:8080/jobs/f47ac10b-58cc-4372-a567-0e02b2c3d479
```

**Expected Response:**
```json
{
  "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "hash": "5d41402abc4b2a76b9719d911017c592ae2e1edc318210e3e8e5e6f8b092c3e2",
  "uploadTime": "42",
  "diskPath": "/mnt/d/firecracker/disks/input-f47ac10b-58cc-4372-a567-0e02b2c3d479.ext4",
  "vmStatus": "ready",
  "scanResult": "pending"
}
```

---

### Test 3: List All Jobs

```bash
curl http://localhost:8080/jobs
```

---

### Test 4: Start Firecracker (Preparation)

Before scanning, start Firecracker daemon:

```bash
# In a separate terminal
sudo rm -f /tmp/firecracker-*.sock

# Start Firecracker for your job (replace {jobID})
sudo firecracker --api-sock /tmp/firecracker-{jobID}.sock
```

---

### Test 5: Launch VM for Scanning

```bash
# Replace {jobID} with your actual job ID
curl -X POST http://localhost:8080/vm/scan/f47ac10b-58cc-4372-a567-0e02b2c3d479
```

**Expected Response:**
```json
{
  "status": "vm_started",
  "jobID": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "socket": "/tmp/firecracker-f47ac10b.sock"
}
```

---

### Test 6: Verify Inside VM (Advanced)

**Connect to VM console:**
```bash
# The VM should boot with the input drive attached as /dev/vdb
# You can verify by checking Firecracker logs or using the API
```

**Check drives inside VM:**
```bash
# If you have console access:
lsblk
# Should show:
# vda - rootfs
# vdb - input drive (your uploaded file)

# Mount and verify file
mkdir -p /mnt/input
mount /dev/vdb /mnt/input
ls -l /mnt/input/
# Should show: input.bin

cat /mnt/input/input.bin
# Should contain: "This is a test file for malware scanning"

# Verify hash inside VM
sha256sum /mnt/input/input.bin
# Should match the original hash
```

---

### Test 7: Cleanup Resources

```bash
# Replace {jobID} with your actual job ID
curl -X DELETE http://localhost:8080/jobs/f47ac10b-58cc-4372-a567-0e02b2c3d479
```

**Expected Response:**
```json
{
  "status": "cleaned",
  "jobID": "f47ac10b-58cc-4372-a567-0e02b2c3d479"
}
```

**Verify cleanup:**
```bash
# Check that files are removed
ls /mnt/d/firecracker/uploads/
ls /mnt/d/firecracker/disks/
# Should not contain files for this jobID

# Check job list
curl http://localhost:8080/jobs
# Should not contain this jobID
```

---

## Complete Workflow Example

```bash
# 1. Start server (in terminal 1)
cd /mnt/d/firecracker
sudo go run *.go

# 2. Upload file (in terminal 2)
echo "Malware sample" > sample.bin
RESPONSE=$(curl -s -F "file=@sample.bin" http://localhost:8080/upload)
echo $RESPONSE

# Extract jobID (using jq if available)
JOB_ID=$(echo $RESPONSE | jq -r '.jobID')
echo "Job ID: $JOB_ID"

# 3. Check job status
curl http://localhost:8080/jobs/$JOB_ID

# 4. Start Firecracker (in terminal 3)
sudo firecracker --api-sock /tmp/firecracker-$JOB_ID.sock

# 5. Launch VM (in terminal 2)
curl -X POST http://localhost:8080/vm/scan/$JOB_ID

# 6. Check updated status
curl http://localhost:8080/jobs/$JOB_ID
# vmStatus should now be "running"

# 7. Cleanup when done
curl -X DELETE http://localhost:8080/jobs/$JOB_ID
```

---

## Troubleshooting

### Error: "Permission denied" on mount
**Solution:** Run server as root: `sudo go run *.go`

### Error: "failed to allocate disk"
**Solution:** Check disk space: `df -h /mnt/d/firecracker/`

### Error: "failed to create ext4 filesystem"
**Solution:** Install e2fsprogs: `sudo apt-get install e2fsprogs`

### Error: "Job not found"
**Solution:** Upload file first before trying to scan

### Error: "Disk image not found for job"
**Solution:** Ensure upload completed successfully before scanning

### Error: Firecracker connection refused
**Solution:** 
1. Start Firecracker with correct socket path
2. Ensure socket permissions: `sudo chmod 666 /tmp/firecracker-*.sock`
3. Check Firecracker is running: `ps aux | grep firecracker`

### Upload file is always same hash
**Solution:** This was fixed! Old version wrote stub content. Current version saves actual file.

---

## Performance Tips

1. **Parallel uploads**: System supports concurrent uploads
2. **Cleanup regularly**: Run cleanup to free disk space
3. **Monitor resources**: `df -h` for disk, `free -h` for memory
4. **VM limits**: Maximum concurrent VMs = RAM / 512MB

---

## Next Steps

1. **Implement scanning logic inside VM**
   - Add analysis scripts to rootfs
   - Extract results via shared folder or API

2. **Add timeout enforcement**
   - Kill VMs after X minutes
   - Automatic cleanup on timeout

3. **Improve job tracking**
   - Use database instead of memory
   - Persist across server restarts

4. **Add authentication**
   - API keys for uploads
   - Rate limiting per client

5. **Set up monitoring**
   - Prometheus metrics
   - Grafana dashboards

---

## Additional Resources

- **Firecracker Documentation**: https://github.com/firecracker-microvm/firecracker/blob/main/docs/getting-started.md
- **API Reference**: See ARCHITECTURE.md
- **Go Documentation**: https://golang.org/doc/

---

## Support

For issues or questions:
1. Check server logs for detailed error messages
2. Review ARCHITECTURE.md for system design
3. Verify all prerequisites are installed
4. Ensure running as root for mount operations
