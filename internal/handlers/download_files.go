package handlers

import "github.com/gofiber/fiber/v2"

func DownloadFile(c *fiber.Ctx) error {
	// Tentukan file path-nya (sesuaikan dengan struktur project kamu)
	filePath := "./files/E Metting.postman_collection.json"

	// Mengirim file dengan nama file download-nya
	// Fiber menyediakan fungsi Download yang mengatur header Content-Disposition secara otomatis.
	return c.Download(filePath, "collection.postman_collection.json")
}
