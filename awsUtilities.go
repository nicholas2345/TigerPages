// All the AWS related functions of our code

package main

import (
	"fmt"
	"mime/multipart"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// Image uploads will happen via POST forms
func uploadImage(file multipart.File, filename string) {

	// create a uploader
	uploader := s3manager.NewUploader(awsSession)

	//// upload the file
	_, err := uploader.Upload(&s3manager.UploadInput{
		// Bucket is where we want to put it
		Bucket: aws.String(os.Getenv("S3_BUCKET")),
		// Key is the filename of what we want
		Key: aws.String(filename),
		// Body is the actual file itself
		Body: file,
		// This makes the object public readable
		ACL: aws.String("public-read"),
	})
	if err != nil {
		fmt.Printf("failed to upload file, %v", err)
	}
}

// deletes a photo with specified filepath
func deletePhoto(filename string) {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(os.Getenv("S3_BUCKET")),
		Key:    aws.String(filename),
	}
	deleter := s3.New(awsSession)
	deleter.DeleteObject(input)
}
