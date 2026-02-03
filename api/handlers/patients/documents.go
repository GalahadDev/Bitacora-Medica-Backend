package patients

import (
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"bitacora-medica-backend/api/config"
	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"
	"bitacora-medica-backend/api/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func UploadDocumentHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		patientIDStr := c.Param("id")
		slog.Info("UploadDocumentHandler Request", "patientID", patientIDStr)
		patientID, err := uuid.Parse(patientIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid patient ID"})
			return
		}

		name := c.PostForm("name")
		categoryStr := c.PostForm("category")
		dateStr := c.PostForm("date")
		description := c.PostForm("description")

		if name == "" || categoryStr == "" || dateStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields (name, category, date)"})
			return
		}

		var category domains.DocumentCategory
		switch domains.DocumentCategory(categoryStr) {
		case domains.DocCategoryLab, domains.DocCategoryImaging, domains.DocCategoryPrescription, domains.DocCategoryOther:
			category = domains.DocumentCategory(categoryStr)
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category. Options: LAB, IMAGING, PRESCRIPTION, OTHER"})
			return
		}

		docDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format (YYYY-MM-DD)"})
			return
		}

		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
			return
		}

		storageSvc := services.NewStorageService(cfg)
		filePath, signedURL, err := storageSvc.UploadPatientDocument(patientIDStr, file)
		if err != nil {
			slog.Error("UploadPatientDocument Failed", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file: " + err.Error()})
			return
		}
		slog.Info("File allocated in Storage", "path", filePath, "signedURL", signedURL)

		doc := domains.PatientDocument{
			PatientID:   patientID,
			Name:        name,
			Category:    category,
			Date:        docDate,
			FileUrl:     filePath, // Guardamos el PATH en lugar de la URL pública
			FileType:    strings.ReplaceAll(filepath.Ext(file.Filename), ".", ""),
			Description: description,
		}

		if err := database.GetDB().Create(&doc).Error; err != nil {
			slog.Error("Failed to save document record", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save document record"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":    "Document uploaded successfully",
			"data":       doc,
			"signed_url": signedURL,
		})
	}
}

func ListDocumentsHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		patientIDStr := c.Param("id")
		if _, err := uuid.Parse(patientIDStr); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid patient ID"})
			return
		}

		var docs []domains.PatientDocument
		result := database.GetDB().Where("patient_id = ?", patientIDStr).Order("date desc").Find(&docs)

		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list documents"})
			return
		}

		storageSvc := services.NewStorageService(cfg)
		for i := range docs {

			if len(docs[i].FileUrl) > 4 && docs[i].FileUrl[:4] == "http" {
				continue
			}
			signed, err := storageSvc.GetPublicURL("patient-documents", docs[i].FileUrl)
			if err == nil {
				docs[i].FileUrl = signed
			}

		}

		c.JSON(http.StatusOK, gin.H{"data": docs})
	}
}

func DeleteDocumentHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		docIDStr := c.Param("doc_id")
		docID, err := uuid.Parse(docIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
			return
		}

		var doc domains.PatientDocument
		if err := database.GetDB().First(&doc, "id = ?", docID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
			return
		}

		if err := database.GetDB().Delete(&doc).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete document"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Document deleted successfully"})
	}
}
