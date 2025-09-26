package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
)

// ReducerService handles the reduction phase of MapReduce
type ReducerService struct {
	s3Client *s3.S3
	bucket   string
}

// ReduceResponse is returned by the reducer API
type ReduceResponse struct {
	Words       map[string]int `json:"words"`
	TotalWords  int            `json:"total_words"`
	UniqueWords int            `json:"unique_words"`
	Message     string         `json:"message"`
}

// MapperResult represents the structure of mapper output
type MapperResult struct {
	Words map[string]int `json:"words"`
	Total int            `json:"total_words"`
}

// FinalResult is the final aggregated word count
type FinalResult struct {
	Words       map[string]int `json:"words"`
	TotalWords  int            `json:"total_words"`
	UniqueWords int            `json:"unique_words"`
}

// downloadFromS3 downloads a file from S3 and returns its content
func (r *ReducerService) downloadFromS3(bucket, key string) ([]byte, error) {
	result, err := r.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	return io.ReadAll(result.Body)
}

// listMapperResults lists all mapper result files in S3
func (r *ReducerService) listMapperResults() ([]string, error) {
	result, err := r.s3Client.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(r.bucket),
		Prefix: aws.String("map_results/"),
	})
	if err != nil {
		return nil, err
	}

	var keys []string
	for _, obj := range result.Contents {
		// Only include JSON files
		if strings.HasSuffix(*obj.Key, ".json") {
			keys = append(keys, *obj.Key)
		}
	}

	return keys, nil
}

// aggregateResults combines all mapper results into final word count
func (r *ReducerService) aggregateResults(mapperKeys []string) (*FinalResult, error) {
	finalWords := make(map[string]int)
	totalWords := 0

	for _, key := range mapperKeys {
		// Download mapper result file
		data, err := r.downloadFromS3(r.bucket, key)
		if err != nil {
			log.Printf("Error downloading %s: %v", key, err)
			continue
		}

		// Parse JSON
		var mapperResult MapperResult
		if err := json.Unmarshal(data, &mapperResult); err != nil {
			log.Printf("Error parsing JSON from %s: %v", key, err)
			continue
		}

		// Aggregate word counts
		for word, count := range mapperResult.Words {
			finalWords[word] += count
		}

		totalWords += mapperResult.Total
	}

	return &FinalResult{
		Words:       finalWords,
		TotalWords:  totalWords,
		UniqueWords: len(finalWords),
	}, nil
}

// reduceHandler handles the HTTP request to reduce all mapper results
func (r *ReducerService) reduceHandler(c *gin.Context) {
	// List all mapper result files
	mapperKeys, err := r.listMapperResults()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to list mapper results: %v", err),
		})
		return
	}

	if len(mapperKeys) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No mapper results found in S3",
		})
		return
	}

	// Aggregate all mapper results
	finalResult, err := r.aggregateResults(mapperKeys)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to aggregate results: %v", err),
		})
		return
	}

	// Convert to JSON and save to S3 for backup/reference
	jsonData, err := json.MarshalIndent(finalResult, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to marshal final result: %v", err),
		})
		return
	}

	// Upload final result to S3
	finalKey := "final_result/word_count_final.json"
	_, err = r.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(finalKey),
		Body:        strings.NewReader(string(jsonData)),
		ContentType: aws.String("application/json"),
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to upload final result: %v", err),
		})
		return
	}

	// Return the actual word count map in the response
	response := ReduceResponse{
		Words:       finalResult.Words,
		TotalWords:  finalResult.TotalWords,
		UniqueWords: finalResult.UniqueWords,
		Message:     fmt.Sprintf("Successfully reduced all mapper results. Total: %d words, Unique: %d words", finalResult.TotalWords, finalResult.UniqueWords),
	}

	c.JSON(http.StatusOK, response)
}

// NewReducerService creates a new reducer service instance
func NewReducerService(bucket string) *ReducerService {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	}))

	return &ReducerService{
		s3Client: s3.New(sess),
		bucket:   bucket,
	}
}

func main() {
	bucket := "mapreduce-experiment-975050147762"

	reducer := NewReducerService(bucket)

	r := gin.Default()
	r.GET("/reduce", reducer.reduceHandler)

	log.Printf("Reducer service starting on port 8080...")
	log.Printf("Using S3 bucket: %s", bucket)
	r.Run(":8080")
}
