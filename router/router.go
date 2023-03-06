package router

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	image_service "example.com/photobank/service"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20 // 8 MiB

	authorized := r.Group("/admin", gin.BasicAuth(gin.Accounts{
		"foo":    "bar",
		"austin": "1234",
		"lena":   "hello2",
		"manu":   "4321",
	}))

	// Upload image
	authorized.POST("/upload", func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.String(http.StatusBadRequest, "get form err: %s", err.Error())
			return
		}
		files := form.File["files"]
		if err := image_service.SaveImage(files, c.SaveUploadedFile); err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		c.String(http.StatusOK, "Uploaded successfully %d files", len(files))
	})

	// Download image by filename 
	authorized.GET("/download/:filename", func(c *gin.Context) {
		filename := c.Param("filename")
		var begin, end int64 
		var err error

 		requestRange := c.Request.Header.Get("range")
		if requestRange == "" {
			begin, end = 0, -1 
			c.Status(http.StatusOK)
		} else {
			requestRange = requestRange[6:] // Strip the "bytes="
			splitRange := strings.Split(requestRange, "-")
			if len(splitRange) != 2 {
				c.String(http.StatusBadRequest, "invalid values for header 'Range'")
				return
			}

			begin, err = strconv.ParseInt(splitRange[0], 10, 64)
			if err != nil {
				c.String(http.StatusBadRequest, "begin string parse err: %s", err.Error())
				return
			}

			end, err = strconv.ParseInt(splitRange[1], 10, 64)
			if splitRange[1] == "" {
				end = -1 
			} else if err != nil {
				c.String(http.StatusBadRequest, "end string parse err: %s", err.Error())
				return
			}
			c.Status(http.StatusPartialContent)
		}

		image, err := image_service.DownloadImage(filename, begin, end)
		defer image.File.Close()
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())	
		}

		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, image.Name))
		c.Header("Content-Type", image.ContentType)
		c.Header("Accept-Ranges", "bytes")
		c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", begin, end, image.NBytes))
		c.Header("Content-Length", strconv.FormatInt(image.NBytes, 10))

		io.CopyN(c.Writer, image.File, image.NBytes)
	})
	
	return r
}
