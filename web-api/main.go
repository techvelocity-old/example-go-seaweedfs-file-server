package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	seaweedfs_master_url = os.Getenv("SEAWEED_MASTER_URL")
)

func getSeaweedfsFidAndUrl() (string, string, error) {
	response, err := http.Get(fmt.Sprintf("%s%s", seaweedfs_master_url, "assign"))
	if err != nil {
		return "", "", fmt.Errorf("error sending GET request: %w", err)
	}
	defer response.Body.Close()

	var data MasterResponse

	err = json.NewDecoder(response.Body).Decode(&data)
	if err != nil {
		return "", "", fmt.Errorf("error parsing JSON: %w", err)
	}

	fid := data.FID
	url := data.URL

	return fid, url, nil
}

func uploadFileToSeaweedfs(fid string, url string, file multipart.File) ([]byte, error) {
	// Create a new buffer to store the multipart request body
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// Add the image file to the multipart form
	part, err := writer.CreateFormFile("file", "image.png")
	if err != nil {
		fmt.Println("Error creating form file:", err)
		return nil, err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		fmt.Println("Error copying file data:", err)
		return nil, err
	}

	// Close the writer to finalize the multipart form
	writer.Close()

	request, err := http.NewRequest("POST", fmt.Sprintf("http://%s/%s", url, fid), body)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	// Set the Content-Type header with the boundary parameter
	request.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Error sending POST request:", err)
		return nil, err
	}
	defer response.Body.Close()

	// Read the response body as a byte array
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil, err
	}

	return responseBody, nil
}

func getSeaweedfsFileLocation() (*Volume, error) {
	response, err := http.Get(fmt.Sprintf("%s%s", seaweedfs_master_url, "lookup?volumeId=3"))
	if err != nil {
		fmt.Println("Error sending GET request:", err)
		return nil, err
	}
	defer response.Body.Close()

	var d Volume

	err = json.NewDecoder(response.Body).Decode(&d)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil, err
	}

	return &d, nil
}

func downloadSeaweedfsFile(d *Volume, fid string) (*bytes.Buffer, error) {
	response, err := http.Get(fmt.Sprintf("http://%s/%s", d.Locations[0].PublicURL, fid))
	if err != nil {
		fmt.Println("Error sending GET request:", err)
		return nil, err
	}
	defer response.Body.Close()

	// Create a buffer to store the content of the response body
	var buf bytes.Buffer

	// Copy the response body to the buffer
	_, err = io.Copy(&buf, response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil, err
	}

	return &buf, nil

}

func getSeaweedfsFileName(db *gorm.DB, fid string) (string, error) {
	file_record := FileRecord{}
	result := db.First(&file_record, "f_id = ?", fid)
	if result.Error != nil {
		log.Printf("Failed to get file record by FID: %v", result.Error)
		return "", result.Error
	}
	return file_record.FileName, nil
}

func main() {
	r := gin.Default()
	db := InitDB()

	r.POST("/api/upload", func(c *gin.Context) {
		// Read the image file from the form data
		file, fh, err := c.Request.FormFile("file")
		if err != nil {
			c.String(http.StatusInternalServerError, "Error reading file: "+err.Error())
			return
		}

		fmt.Println(file)

		fid, url, err := getSeaweedfsFidAndUrl()
		if err != nil {
			c.String(http.StatusInternalServerError, "Error getting SeaweedFS FID and URL: "+err.Error())
			return
		}

		responseBody, err := uploadFileToSeaweedfs(fid, url, file)
		fmt.Println(responseBody)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error uploading file to SeaweedFS: "+err.Error())
			return
		}

		f := FileRecord{FileName: fh.Filename, FID: fid}

		db.Create(&f)

		c.JSON(http.StatusOK, fid)

	})
	r.GET("/api/download/:fid", func(c *gin.Context) {
		fid := c.Param("fid")
		file_location, err := getSeaweedfsFileLocation()
		if err != nil {
			log.Println(err)
			c.String(http.StatusInternalServerError, "Error getting file from SeaweedFS")
			return
		}
		buf, err := downloadSeaweedfsFile(file_location, fid)
		if err != nil {
			log.Println(err)
			c.String(http.StatusInternalServerError, "Error downloading file from SeaweedFS")
			return
		}

		file_name, err := getSeaweedfsFileName(db, fid)
		if err != nil {
			log.Println(err)
			return
		}

		// Set the appropriate headers for file download
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file_name))
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Length", fmt.Sprint(buf.Len()))

		// Write the buffer directly to the response writer
		if _, err := buf.WriteTo(c.Writer); err != nil {
			log.Println(err)
			c.String(http.StatusInternalServerError, "Error writing file to response")
			return
		}
	})

	r.GET("/api/files", func(c *gin.Context) {
		var file_records []FileRecord
		db.Find(&file_records)
		c.JSON(200, file_records)
	})

	r.Run(":8080")
}
