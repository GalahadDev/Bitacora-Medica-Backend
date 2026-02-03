package common

import (
	"net/http"

	"bitacora-medica-backend/api/config"
	"bitacora-medica-backend/api/services"

	"github.com/gin-gonic/gin"
)

func UploadImageHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File is mandatory"})
			return
		}

		patientID := c.PostForm("patient_id")
		if patientID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "patient_id is required"})
			return
		}

		storage := services.NewStorageService(cfg)
		path, signedURL, err := storage.UploadImage(patientID, file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":    "Image uploaded successfully",
			"path":       path,
			"signed_url": signedURL,
			"url":        signedURL,
		})
	}
}

func UploadConsentHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File is mandatory"})
			return
		}

		patientID := c.PostForm("patient_id")
		if patientID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "patient_id is required"})
			return
		}

		storage := services.NewStorageService(cfg)
		path, signedURL, err := storage.UploadConsentPDF(patientID, file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload PDF", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":    "Consent PDF uploaded successfully",
			"path":       path,
			"signed_url": signedURL,
			"url":        signedURL,
		})
	}
}
