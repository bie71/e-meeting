package handlers

import (
	"e_metting/internal/config"
	"e_metting/internal/services/supabase"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type UploadHandler struct {
	uploadService *supabase.Client
	cfg           *config.Config
}

func NewUploadHandler(uploadService *supabase.Client, cfg *config.Config) *UploadHandler {
	return &UploadHandler{uploadService: uploadService, cfg: cfg}
}
func (h *UploadHandler) UploadHandler(c *fiber.Ctx) error {
	// Ambil file dari form
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file is required"})
	}

	safeFilename := sanitizeFilename(fileHeader.Filename)

	// Simpan ke temporary directory OS
	tempPath := filepath.Join(os.TempDir(), safeFilename) // ✅ Simpan file temporer di /tmp/

	if err := c.SaveFile(fileHeader, tempPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save file"})
	}
	defer os.Remove(tempPath)

	objectName := "uploads/" + safeFilename

	err = h.uploadService.UploadToSupabase(objectName, tempPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Public URL jika bucket diatur "public"
	fileURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s",
		h.cfg.Client.Endpoint,
		h.cfg.Client.BucketName,
		objectName,
	)

	return c.JSON(fiber.Map{
		"message": "upload success",
		"url":     fileURL,
	})
}

func sanitizeFilename(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")

	// Hapus karakter aneh, hanya izinkan huruf, angka, titik, dash, underscore
	re := regexp.MustCompile(`[^a-z0-9._-]+`)
	name = re.ReplaceAllString(name, "")

	// Batasi panjang nama file kalau perlu
	if len(name) > 100 {
		name = name[:100]
	}

	return name
}
