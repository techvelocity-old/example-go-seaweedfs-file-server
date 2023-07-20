package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	seaweedfs_master_url = "http://seaweedfs-master:9333/dir/"
)

func get_seaweedfs_fid_and_url() (string, string, error) {
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

func upload_file_to_seaweedfs(fid string, url string, file multipart.File) ([]byte, error) {
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

func get_seaweedfs_file_location() (*Volume, error) {
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

func download_seaweedfs_file(d *Volume, fid string) (*bytes.Buffer, error) {
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

func main() {
	r := gin.Default()
	db := InitDB()
	r.LoadHTMLGlob("templates/*.html")

	index := func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	}

	r.GET("/", index)
	r.POST("/upload", func(c *gin.Context) {
		// Read the image file from the form data
		file, fh, err := c.Request.FormFile("file")
		if err != nil {
			c.String(http.StatusInternalServerError, "Error reading file: "+err.Error())
			return
		}

		fmt.Println(file)

		fid, url, err := get_seaweedfs_fid_and_url()
		if err != nil {
			c.String(http.StatusInternalServerError, "Error getting SeaweedFS FID and URL: "+err.Error())
			return
		}

		responseBody, err := upload_file_to_seaweedfs(fid, url, file)
		fmt.Println(responseBody)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error uploading file to SeaweedFS: "+err.Error())
			return
		}

		f := FileRecord{FileName: fh.Filename, FID: fid}

		db.Create(&f)

		c.JSON(http.StatusOK, fid)

	})
	r.POST("/download", func(c *gin.Context) {
		fid := c.Request.FormValue("fid")
		file_location, err := get_seaweedfs_file_location()
		if err != nil {
			log.Println(err)
			return
		}
		buf, err := download_seaweedfs_file(file_location, fid)
		if err != nil {
			log.Println(err)
			return
		}

		c.Data(http.StatusOK, "application/octet-stream", buf.Bytes())

	})

	r.GET("/files", func(c *gin.Context) {
		var file_records []FileRecord
		db.Find(&file_records)
		c.JSON(200, file_records)
	})

	r.Run(":8080")
}
