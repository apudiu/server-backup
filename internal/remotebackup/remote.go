package remotebackup

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/apudiu/server-backup/internal/util"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"log"
	"os"
	"path/filepath"
)

type UlDl struct {
	client            *s3.Client
	bucket            string
	transferChunkSize int64
}

// BucketExists checks whether bucket exists in the current account.
func (ud *UlDl) BucketExists() (bool, error) {
	_, err := ud.client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
		Bucket: aws.String(ud.bucket),
	})

	exists := true
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			switch apiError.(type) {
			case *types.NotFound:
				log.Printf("Bucket %v is available.\n", ud.bucket)
				exists = false
				err = nil
			default:
				log.Printf("Either you don't have access to bucket %v or another error occurred. "+
					"Here's what happened: %v\n", ud.bucket, err)
			}
		}
	} else {
		log.Printf("Bucket %v exists and you already own it.", ud.bucket)
	}

	return exists, err
}

// DeleteObjects deletes a list of objects from a bucket.
func (ud *UlDl) DeleteObjects(objectKeys []string) (error, []types.DeletedObject) {
	var objectIds []types.ObjectIdentifier
	for _, key := range objectKeys {
		objectIds = append(objectIds, types.ObjectIdentifier{Key: aws.String(key)})
	}
	output, err := ud.client.DeleteObjects(context.TODO(), &s3.DeleteObjectsInput{
		Bucket: aws.String(ud.bucket),
		Delete: &types.Delete{Objects: objectIds},
	})
	if err != nil {
		log.Printf("Couldn't delete objects from bucket %v. Here's why: %v\n", ud.bucket, err)
	} else {
		log.Printf("Deleted %v objects.\n", len(output.Deleted))
	}
	return err, output.Deleted
}

// ListObjects lists the objects in bucket.
func (ud *UlDl) ListObjects() ([]types.Object, error) {
	result, err := ud.client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(ud.bucket),
	})
	var contents []types.Object
	if err != nil {
		log.Printf("Couldn't list objects in bucket %v. Here's why: %v\n", ud.bucket, err)
	} else {
		contents = result.Contents
	}
	return contents, err
}

// CopyToFolder copies an object in a bucket to a sub folder in the same bucket.
func (ud *UlDl) CopyToFolder(objectKey string, folderName string) error {
	_, err := ud.client.CopyObject(context.TODO(), &s3.CopyObjectInput{
		Bucket:     aws.String(ud.bucket),
		CopySource: aws.String(fmt.Sprintf("%v/%v", ud.bucket, objectKey)),
		Key:        aws.String(fmt.Sprintf("%v/%v", folderName, objectKey)),
	})
	if err != nil {
		log.Printf("Couldn't copy object from %v:%v to %v:%v/%v. Here's why: %v\n",
			ud.bucket, objectKey, ud.bucket, folderName, objectKey, err)
	}
	return err
}

// UploadObject uses an upload manager to upload data to an object in a bucket.
// The upload manager breaks large data into parts and uploads the parts concurrently.
func (ud *UlDl) UploadObject(objectKey string, largeObject []byte) (uploadResult *manager.UploadOutput, err error) {
	largeBuffer := bytes.NewReader(largeObject)

	uploader := manager.NewUploader(ud.client, func(u *manager.Uploader) {
		u.PartSize = ud.transferChunkSize
	})
	uploadResult, err = uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(ud.bucket),
		Key:    aws.String(objectKey),
		Body:   largeBuffer,
	})
	if err != nil {
		log.Printf("Couldn't upload large object to %v:%v. Here's why: %v\n",
			ud.bucket, objectKey, err)
		return nil, err
	}

	return uploadResult, nil
}

// DownloadObject uses a download manager to download an object from a bucket.
// The download manager gets the data in parts and writes them to a buffer until all of
// the data has been downloaded.
func (ud *UlDl) DownloadObject(objectKey string) ([]byte, error) {
	downloader := manager.NewDownloader(ud.client, func(d *manager.Downloader) {
		d.PartSize = ud.transferChunkSize
	})
	buffer := manager.NewWriteAtBuffer([]byte{})
	_, err := downloader.Download(context.TODO(), buffer, &s3.GetObjectInput{
		Bucket: aws.String(ud.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		log.Printf("Couldn't download large object from %v:%v. Here's why: %v\n",
			ud.bucket, objectKey, err)
	}
	return buffer.Bytes(), err
}

// DownloadToFile uses a download manager to download an object from a bucket to a file
func (ud *UlDl) DownloadToFile(objectKey string, targetDirectory string) error {
	// Create the directories in the path
	file := filepath.Join(targetDirectory, objectKey)
	if err := os.MkdirAll(filepath.Dir(file), 0775); err != nil {
		return err
	}

	// Set up the local file
	fd, err := os.Create(file)
	if err != nil {
		return err
	}
	//goland:noinspection ALL
	defer fd.Close()

	downloader := manager.NewDownloader(ud.client, func(d *manager.Downloader) {
		d.PartSize = ud.transferChunkSize
	})

	_, err = downloader.Download(context.TODO(), fd, &s3.GetObjectInput{
		Bucket: aws.String(ud.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		log.Printf("Couldn't download large object from %v:%v. Here's why: %v\n",
			ud.bucket, objectKey, err)
	}
	return err
}

func New(user, bucket string, transferChunkSizeMb uint8) (*UlDl, error) {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithSharedConfigProfile(user),
	)
	if err != nil {
		return nil, err
	}

	// Create an Amazon S3 service client
	client := s3.NewFromConfig(cfg)
	return &UlDl{
		client:            client,
		bucket:            bucket,
		transferChunkSize: util.GetBytesForMb(int64(transferChunkSizeMb)),
	}, nil
}
