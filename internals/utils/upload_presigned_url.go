package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

func UploadToPresignedURL(method string, presignedURL string, file multipart.File, fileHeader *multipart.FileHeader) error {

	if method == "PUT" {

		req, err := http.NewRequest(method, presignedURL, file)
		if err != nil {
			return fmt.Errorf("new request: %w", err)
		}

		// Content-Length wajib di-set biar server tahu ukuran body
		req.ContentLength = fileHeader.Size

		// Content-Type harus SAMA PERSIS dengan yang dipakai saat generate presigned URL
		// (kalau presigned URL di-generate dengan ContentType tertentu, harus cocok)
		contentType := fileHeader.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "text/csv"
		}
		req.Header.Set("Content-Type", contentType)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("do request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("upload failed, status %d: %s", resp.StatusCode, string(body))
		}

		return nil
	} else {
		return fmt.Errorf("Error Method type: %s. Must be PUT", method)
	}

}
