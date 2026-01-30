package controllers

import (
	"cursy_back/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UploadImage handles file uploads
func UploadImage(c *gin.Context) {
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file content"})
		return
	}
	defer file.Close()

	// Initialize storage service
	storageService := services.NewStorageService()
	if storageService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Storage service not available"})
		return
	}

	url, err := storageService.UploadFile(file, fileHeader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}
