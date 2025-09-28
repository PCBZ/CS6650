package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
)

type SplitterService struct {
	s3Client *s3.S3
	bucket   string
}

type SplitResponse struct {
	ChunkURLs []string `json:"chunk_urls"`
	Message   string   `json:"message"`
}

func (s *SplitterService) downloadFile(fileURL string) (io.ReadCloser, error) {
	if strings.Contains(fileURL, "s3://") {
		// Handle S3 URL
		parts := strings.SplitN(fileURL[5:], "/", 2) // Remove "s3://" prefix

		bucket := parts[0]
		key := parts[1]

		result, err := s.s3Client.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to download from S3: %v", err)
		}
		return result.Body, nil
	}
	// Handle HTTPS URL
	response, err := http.Get(fileURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download from HTTPS: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		response.Body.Close()
		return nil, fmt.Errorf("HTTPS request failed with status: %d", response.StatusCode)
	}
	return response.Body, nil
}

func NewSplitterService(bucket string) *SplitterService {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	}))

	return &SplitterService{
		s3Client: s3.New(sess),
		bucket:   bucket,
	}
}

func (s *SplitterService) splitHandler(c *gin.Context) {
	// Get the HTTPS URL parameter
	fileURL := c.Query("url")
	if fileURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing 'url' query parameter",
		})
		return
	}
	// Get the number of chunks parameter
	chunksStr := c.DefaultQuery("chunks", "3")
	numChunks, err := strconv.Atoi(chunksStr)
	if err != nil || numChunks < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid 'chunks' parameter. Must be a number larger than 0",
		})
		return
	}

	// Download the file from URL
	reader, err := s.downloadFile(fileURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to download file: %v", err),
		})
		return
	}
	defer reader.Close()

	// Initialize chunk buffers, each string builder writes lines to its own chunk file
	chunkBuffers := make([]*strings.Builder, numChunks)
	for i := range chunkBuffers {
		chunkBuffers[i] = &strings.Builder{}
	}

	scanner := bufio.NewScanner(reader)
	lineIdx := 0
	totalLines := 0
	for scanner.Scan() {
		// read each line
		line := scanner.Text()
		// according to line index, dispatch to different chunk buffer.
		chunkIdx := lineIdx % numChunks // round-robin
		chunkBuffers[chunkIdx].WriteString(line)
		chunkBuffers[chunkIdx].WriteString("\n")
		lineIdx++
		totalLines++
	}
	if err := scanner.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Error reading file: %v", err),
		})
		return
	}
	if totalLines == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File is empty",
		})
		return
	}

	// Upload chunks to S3 and collect URLs
	var chunkURLs []string
	for i, buf := range chunkBuffers {
		chunkContent := buf.String()
		chunkKey := fmt.Sprintf("chunks/chunk_%d.txt", i+1)
		_, err := s.s3Client.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(chunkKey),
			Body:   strings.NewReader(chunkContent),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to upload chunk %d: %v", i+1, err),
			})
			return
		}
		chunkURL := fmt.Sprintf("s3://%s/%s", s.bucket, chunkKey)
		chunkURLs = append(chunkURLs, chunkURL)
	}

	// Return the chunk URLs
	response := SplitResponse{
		ChunkURLs: chunkURLs,
		Message:   fmt.Sprintf("Successfully split file into %d chunks with %d total lines", len(chunkBuffers), totalLines),
	}

	c.JSON(http.StatusOK, response)
}

func main() {
	// Get bucket name from environment variable or use default
	bucket := "mapreduce-experiment-975050147762" // Replace with your bucket name

	splitter := NewSplitterService(bucket)

	r := gin.Default()

	// Routes
	r.GET("/split", splitter.splitHandler)

	// Start server
	log.Printf("Splitter service starting on port 8080...")
	log.Printf("Using S3 bucket: %s", bucket)
	r.Run(":8080")
}
