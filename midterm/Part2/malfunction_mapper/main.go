package main

import (
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
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "Simulated malfunction: mapper failed intentionally",
	})
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
