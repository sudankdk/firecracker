package handler

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/sudankdk/firecracker/internal/sandboxing"
)

type UploadHandler struct {
	VM *sandboxing.VMManager
}

func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. Parse file
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("upload rejected: missing file part: %v", err)
		http.Error(w, "file required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 2. Save upload to disk
	uploadID := uuid.New().String()
	uploadPath := filepath.Join(h.VM.BaseUploadDir, uploadID)
	log.Printf("upload received: id=%s name=%s", uploadID, header.Filename)

	out, err := os.Create(uploadPath)
	if err != nil {
		log.Printf("upload storage failed: id=%s path=%s err=%v", uploadID, uploadPath, err)
		http.Error(w, "cannot save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	bytesWritten, err := io.Copy(out, file)
	if err != nil {
		log.Printf("upload write failed: id=%s err=%v", uploadID, err)
		http.Error(w, "upload failed", http.StatusInternalServerError)
		return
	}
	log.Printf("upload stored: id=%s path=%s bytes=%d", uploadID, uploadPath, bytesWritten)

	// 3. Spawn sandbox VM
	vm, err := h.VM.SpawnVM(uploadPath)
	if err != nil {
		log.Printf("sandbox launch failed: upload=%s err=%v", uploadID, err)
		http.Error(w, "sandbox failed", http.StatusInternalServerError)
		return
	}
	log.Printf("sandbox running: upload=%s vm=%s", uploadID, vm.ID)

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("file accepted and sandboxed"))
}
