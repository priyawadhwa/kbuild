package storage

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	// "golang.org/x/oauth2/google"
	"io"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
)

// CreateStorageBucket creates a storage bucket to store the source context in
func CreateStorageBucket(sc string) (string, error) {
	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", nil
	}
	projectID, err := getProjectID("")
	bucketName := fmt.Sprintf("kbuild-buckets-%d", time.Now().Unix())
	// Creates a Bucket instance.
	bucket := client.Bucket(bucketName)

	// Creates the new bucket.
	if err := bucket.Create(ctx, projectID, nil); err != nil {
		logrus.Errorf("Failed to create bucket: %v", err)
		return "", err
	}
	logrus.Info("Created bucket ", bucketName)
	return bucketName, nil
}

// DeleteStorageBucket deletes the storage bucket the source context is in
func DeleteStorageBucket() error {

	return nil
}

// UploadSourceContextToStorageBucket uploads the source context to the storage bucket
func UploadSourceContextToStorageBucket(sc string, bucket *storage.BucketHandle) error {
	return filepath.Walk(sc, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		f, err := os.Open(info.Name())
		if err != nil {
			logrus.Debugf("Could not open", info.Name())
			return nil
		}
		defer f.Close()
		buf := bytes.NewBuffer(nil)
		_, err = io.Copy(buf, f)
		if err != nil {
			logrus.Debugf("Could not copy contents of", info.Name())
			return nil
		}
		return uploadFile(bucket, buf.Bytes(), path)
	})
}

// uploadFile uploads a file to a Google Cloud Storage bucket.
func uploadFile(bucket *storage.BucketHandle, fileContents []byte, path string) error {
	ctx := context.Background()
	obj := bucket.Object(path)
	// Write something to obj.
	// w implements io.Writer.
	w := obj.NewWriter(ctx)
	// Close, just like writing a file.
	if err := w.Close(); err != nil {
		logrus.Debug("Failed to close file")
	}
	if _, err := w.Write(fileContents); err != nil {
		logrus.Error("createFile: unable to write file %q: %v", path, err)
		return err
	}
	if err := w.Close(); err != nil {
		logrus.Error("createFile: unable to close bucket: %v", err)
		return err
	}
	return nil
}

// GetFilesFromStorageBucket gets files/dirs that match the context
func GetFilesFromStorageBucket(context string) error {

	return nil
}

func getProjectID(scope string) (string, error) {
	// ctx := context.Background()
	// // defaultCreds, err := google.FindDefaultCredentials(ctx, scope)
	// if err != nil {
	// 	return "", err
	// }
	// return defaultCreds.ProjectID, nil
	return "priya-wadhwa", nil
}
