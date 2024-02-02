package remotebackup

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"log"
	"os"
	"path/filepath"
)

var bucketName = "all-server-backup-bucket"

func ListBucket() {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithSharedConfigProfile("server-backup"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create an Amazon S3 service client
	client := s3.NewFromConfig(cfg)

	downloader := manager.NewDownloader(client)

	// Get the first page of results for ListObjectsV2 for a bucket
	output, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("first page results:")
	for _, object := range output.Contents {
		log.Printf("key=%s size=%d", aws.ToString(object.Key), object.Size)
	}

	//

	dlFile := output.Contents[1].Key
	resp, err := client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    dlFile,
	})
	if err != nil {
		fmt.Println("File download err for", dlFile, err.Error())
	}
	defer resp.Body.Close()

	// todo: upload a file
	dlErr := downloadToFile(downloader, "./downloads", bucketName, *dlFile)
	if dlErr != nil {
		fmt.Println("File download err", dlErr.Error())
	}
}

func downloadToFile(downloader *manager.Downloader, targetDirectory, bucket, key string) error {
	// Create the directories in the path
	file := filepath.Join(targetDirectory, key)
	if err := os.MkdirAll(filepath.Dir(file), 0775); err != nil {
		return err
	}

	// Set up the local file
	fd, err := os.Create(file)
	if err != nil {
		return err
	}
	defer fd.Close()

	// Download the file using the AWS SDK for Go
	fmt.Printf("Downloading s3://%s/%s to %s...\n", bucket, key, file)
	_, err = downloader.Download(context.TODO(), fd, &s3.GetObjectInput{Bucket: &bucket, Key: &key})

	return err
}
