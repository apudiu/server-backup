package remotebackup

import (
	"context"
	"errors"
	"fmt"
	"github.com/apudiu/server-backup/internal/logger"
	"github.com/apudiu/server-backup/internal/util"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

type UlDl struct {
	client            *s3.Client
	bucket, localDir  string
	transferChunkSize int64
	logger            *logger.Logger
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
				ud.logger.AddHeader(
					util.ServerFailLogf("Bucket %v is unavailable", ud.bucket),
				)
				exists = false
				err = nil
			default:
				ud.logger.AddHeader(
					util.ServerFailLogf("Bucket %s access error. %s", ud.bucket, err.Error()),
				)
			}
		}
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
		ud.logger.AddHeader(
			util.ServerFailLogf("Couldn't delete objects from bucket %s. Here's why: %s", ud.bucket, err.Error()),
		)
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
		ud.logger.AddHeader(
			util.ServerFailLogf("Couldn't list objects in bucket %s. Here's why: %s", ud.bucket, err.Error()),
		)
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
		ud.logger.AddHeader(
			util.ServerFailLogf(
				"Couldn't copy object from %s:%s to %s:%s/%s. Here's why: %s",
				ud.bucket, objectKey, ud.bucket, folderName, objectKey, err.Error(),
			),
		)
	}
	return err
}

// UploadObject uses an upload manager to upload data to an object in a bucket.
// The upload manager breaks large data into parts and uploads the parts concurrently.
func (ud *UlDl) UploadObject(objectKey string, file io.Reader) (uploadResult *manager.UploadOutput, err error) {
	uploader := manager.NewUploader(ud.client, func(u *manager.Uploader) {
		u.PartSize = ud.transferChunkSize
	})
	uploadResult, err = uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(ud.bucket),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	if err != nil {
		ud.logger.AddHeader(
			util.ServerFailLogf(
				"Couldn't upload large object to %s:%s. Here's why: %s",
				ud.bucket, objectKey, err.Error(),
			),
		)
		return nil, err
	}

	return uploadResult, nil
}

// DownloadObject uses a download manager to download an object from a bucket.
// The download manager gets the data in parts and writes them to a buffer until all
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
		ud.logger.AddHeader(
			util.ServerFailLogf(
				"Couldn't download large object from %s:%s. Here's why: %s",
				ud.bucket, objectKey, err.Error(),
			),
		)
	}
	return err
}

// UploadChangedOrNew uploads changed or newly added files to cloud from local backup dir
func (ud *UlDl) UploadChangedOrNew() error {
	// get remote contents
	remoteContents, remoteErr := ud.ListObjects()
	if remoteErr != nil {
		return remoteErr
	}

	// make remote contents map
	rcMap := make(map[string]int64)

	for _, rc := range remoteContents {
		// skip (only) dirs
		if *rc.Size < 1 {
			continue
		}

		rcMap[*rc.Key] = *rc.Size
	}

	// for upload
	var fileList []string

	// traverse local backup dir
	walkEr := filepath.WalkDir(ud.localDir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return err
		}

		fi, err := d.Info()
		if err != nil {
			return err
		}

		// check if this file exist in remote
		rfSize, found := rcMap[path]

		// when missing in remote, add to upload list
		if !found {
			fileList = append(fileList, path)
			return err
		}

		// when found compare the size & if differ upload new ver
		if rfSize != fi.Size() {
			fileList = append(fileList, path)
		} else {
			ud.logger.AddHeader(
				util.ServerLogf("Skipping: %s", path),
			)
		}

		return err
	})

	if walkEr != nil {
		return walkEr
	}

	// perform upload
	for _, list := range fileList {
		// using fn here for closing the file immediately after we're done with it
		// else no file will be closed before all files are done
		func(fp string) {
			f, e := os.OpenFile(fp, os.O_RDONLY, 0644)
			if e != nil {
				ud.logger.AddHeader(
					util.ServerFailLogf("%s error %s", fp, e.Error()),
				)
				return
			}
			defer f.Close()

			ud.logger.AddHeader(
				util.ServerLogf("Uploading: %s", fp),
			)
			_, upErr := ud.UploadObject(fp, f)
			if upErr != nil {
				ud.logger.AddHeader(
					util.ServerFailLogf("Upload err: %s", fp),
				)
			}
		}(list)
	}

	ud.logger.AddHeader(
		util.ServerFailLogf("Uploaded %d files", len(fileList)),
	)

	return nil
}

func New(
	user, bucket, localBackupDir string,
	transferChunkSizeMb uint8,
	l *logger.Logger,
) (*UlDl, error) {
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
		localDir:          localBackupDir,
		transferChunkSize: util.GetBytesForMb(int64(transferChunkSizeMb)),
		logger:            l,
	}, nil
}
