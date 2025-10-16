package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
)

// MapperService holds the S3 client and bucket configuration
type MapperService struct {
	s3Client *s3.S3
	bucket   string
}

// MapResponse is the JSON response returned by the mapper
type MapResponse struct {
	ResultURL string `json:"result_url"`
	WordCount int    `json:"word_count"`
	Message   string `json:"message"`
}

// WordCountResult is the structure stored in S3 with word frequencies
type WordCountResult struct {
	Words map[string]int `json:"words"`
	Total int            `json:"total_words"`
}

// downloads a file from S3 given an S3 URL
func (m *MapperService) downloadFromS3(s3URL string) (io.ReadCloser, error) {
	parts := strings.SplitN(s3URL[5:], "/", 2) // Remove "s3://" prefix
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid S3 URL format")
	}
	bucket := parts[0]
	key := parts[1]

	result, err := m.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	return result.Body, nil
}

// countText counts word frequencies in the given text
func (m *MapperService) countText(text string) map[string]int {
	wordCount := make(map[string]int)

	// Regular expression to match letters, numbers, and apostrophes
	// This handles contractions like 'twere, don't, etc.
	re := regexp.MustCompile(`[a-zA-Z0-9']+`)
	words := re.FindAllString(text, -1)

	for _, word := range words {
		// Convert to lowercase and remove leading/trailing apostrophes
		word = strings.ToLower(strings.Trim(word, "'"))
		if word != "" {
			wordCount[word]++
		}
	}

	return wordCount
}

// mapHandler handles the HTTP request to count a text chunk
func (m *MapperService) mapHandler(c *gin.Context) {
	// Get the chunk S3 URL from query parameters
	chunkURL := c.Query("url")
	if chunkURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing 'url' parameter",
		})
		return
	}

	// Download the chunk from S3
	reader, err := m.downloadFromS3(chunkURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to download chunk: %v", err),
		})
		return
	}
	defer reader.Close()

	// Count word frequencies line by line
	wordCount := make(map[string]int)
	re := regexp.MustCompile(`[a-zA-Z0-9']+`)
	scanner := bufio.NewScanner(reader)
	totalWords := 0
	for scanner.Scan() {
		line := scanner.Text()
		words := re.FindAllString(line, -1)
		for _, word := range words {
			word = strings.ToLower(strings.Trim(word, "'"))
			if word != "" {
				wordCount[word]++
				totalWords++
			}
		}
	}
	if err := scanner.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Error reading chunk: %v", err),
		})
		return
	}

	// Prepare result structure
	result := WordCountResult{
		Words: wordCount,
		Total: totalWords,
	}

	// Convert result to JSON
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to marshal JSON: %v", err),
		})
		return
	}

	// Generate result file name based on input chunk name
	chunkParts := strings.Split(chunkURL, "/")
	chunkName := chunkParts[len(chunkParts)-1]
	resultKey := fmt.Sprintf("map_results/result_%s.json", strings.TrimSuffix(chunkName, ".txt"))

	// Upload result to S3
	_, err = m.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(m.bucket),
		Key:         aws.String(resultKey),
		Body:        strings.NewReader(string(jsonData)),
		ContentType: aws.String("application/json"),
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to upload result: %v", err),
		})
		return
	}

	// Return success response
	resultURL := fmt.Sprintf("s3://%s/%s", m.bucket, resultKey)
	response := MapResponse{
		ResultURL: resultURL,
		WordCount: len(wordCount),
		Message:   fmt.Sprintf("Successfully processed chunk. Found %d unique words, %d total words", len(wordCount), totalWords),
	}

	c.JSON(http.StatusOK, response)
}

// NewMapperService creates a new mapper service instance
func NewMapperService(bucket string) *MapperService {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	}))

	return &MapperService{
		s3Client: s3.New(sess),
		bucket:   bucket,
	}
}

func main() {
	bucket := "mapreduce-experiment-975050147762"

	mapper := NewMapperService(bucket)

	r := gin.Default()
	r.GET("/map", mapper.mapHandler)

	log.Printf("Mapper service starting on port 8080...")
	log.Printf("Using S3 bucket: %s", bucket)
	r.Run(":8080")
}
