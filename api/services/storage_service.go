package services

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	"bitacora-medica-backend/api/config"

	"github.com/google/uuid"
)

type StorageService struct {
	Config *config.Config
}

func NewStorageService(cfg *config.Config) *StorageService {
	return &StorageService{Config: cfg}
}

func (s *StorageService) GetPublicURL(bucket, path string) (string, error) {

	url := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.Config.SupabaseURL, bucket, path)
	return url, nil
}

func (s *StorageService) UploadConsentPDF(patientID string, file *multipart.FileHeader) (path string, signedURL string, err error) {

	ext := filepath.Ext(file.Filename)
	if ext != ".pdf" {
		return "", "", fmt.Errorf("only PDF files are allowed")
	}

	src, err := file.Open()
	if err != nil {
		return "", "", err
	}
	defer src.Close()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return "", "", err
	}

	fileName := fmt.Sprintf("consent_%d.pdf", time.Now().Unix())
	filePath := fmt.Sprintf("%s/%s", patientID, fileName)
	bucketName := "patients-consent"

	err = s.uploadToSupabase(bucketName, filePath, fileBytes, "application/pdf")
	if err != nil {
		return "", "", err
	}

	signedURL, err = s.GetPublicURL(bucketName, filePath)
	return filePath, signedURL, err
}

func (s *StorageService) UploadImage(patientID string, file *multipart.FileHeader) (path string, signedURL string, err error) {

	ext := filepath.Ext(file.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return "", "", fmt.Errorf("only JPG/PNG images are allowed")
	}

	src, err := file.Open()
	if err != nil {
		return "", "", err
	}
	defer src.Close()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return "", "", err
	}

	fileName := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
	filePath := fmt.Sprintf("%s/%s", patientID, fileName)
	bucketName := "session-evidence"

	contentType := http.DetectContentType(fileBytes)
	err = s.uploadToSupabase(bucketName, filePath, fileBytes, contentType)
	if err != nil {
		return "", "", err
	}

	signedURL, err = s.GetPublicURL(bucketName, filePath)
	return filePath, signedURL, err
}

func (s *StorageService) UploadPatientDocument(patientID string, file *multipart.FileHeader) (path string, signedURL string, err error) {

	ext := filepath.Ext(file.Filename)
	validExts := map[string]bool{".pdf": true, ".jpg": true, ".jpeg": true, ".png": true}
	if !validExts[ext] {
		return "", "", fmt.Errorf("unsupported file type: %s", ext)
	}

	src, err := file.Open()
	if err != nil {
		return "", "", err
	}
	defer src.Close()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return "", "", err
	}

	fileName := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
	filePath := fmt.Sprintf("%s/%s", patientID, fileName)
	bucketName := "patient-documents"

	contentType := http.DetectContentType(fileBytes)
	if ext == ".pdf" {
		contentType = "application/pdf"
	}
	err = s.uploadToSupabase(bucketName, filePath, fileBytes, contentType)
	if err != nil {
		return "", "", err
	}

	signedURL, err = s.GetPublicURL(bucketName, filePath)
	return filePath, signedURL, err
}

func (s *StorageService) uploadToSupabase(bucket, path string, data []byte, contentType string) error {
	slog.Info("Uploading to Supabase", "bucket", bucket, "path", path, "size", len(data), "contentType", contentType)

	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.Config.SupabaseURL, bucket, path)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+s.Config.SupabaseKey)
	req.Header.Set("Content-Type", contentType)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Supabase request failed", "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("upload failed with status: %d", resp.StatusCode)
	}
	return nil
}
