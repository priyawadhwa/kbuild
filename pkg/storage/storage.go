package storage

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	// "golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"io"
	"io/ioutil"
	"os"
	"time"

	"cloud.google.com/go/storage"
)

// CreateStorageBucket creates a storage bucket to store the source context in
func CreateStorageBucket() (*storage.BucketHandle, string, error) {
	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, "", err
	}
	projectID, err := getProjectID("")
	bucketName := fmt.Sprintf("kbuild-buckets-%d", time.Now().Unix())
	// Creates a Bucket instance.
	bucket := client.Bucket(bucketName)

	// Creates the new bucket.
	if err := bucket.Create(ctx, projectID, nil); err != nil {
		logrus.Errorf("Failed to create bucket: %v", err)
		return nil, "", err
	}
	logrus.Info("Created bucket ", bucketName)
	return bucket, bucketName, nil
}

// DeleteStorageBucket deletes the storage bucket
func DeleteStorageBucket(bucketName string) error {
	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	bucket := client.Bucket(bucketName)
	bucket.Delete(ctx)
	return nil
}

// UploadContextToBucket uploads the given context to the given bucket
func UploadContextToBucket(files []string, bucket *storage.BucketHandle) error {
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			logrus.Debugf("Could not open %s, err: %v", file, err)
			return nil
		}
		defer f.Close()
		buf := bytes.NewBuffer(nil)
		_, err = io.Copy(buf, f)
		if err != nil {
			logrus.Debugf("Could not copy contents of %s, err: %v", file, err)
			return nil
		}
		if err := uploadFile(bucket, buf.Bytes(), file); err != nil {
			return err
		}
	}
	return nil
}

// uploadFile uploads a file to a Google Cloud Storage bucket.
func uploadFile(bucket *storage.BucketHandle, fileContents []byte, path string) error {
	ctx := context.Background()
	// Write something to obj.
	// w implements io.Writer.
	w := bucket.Object(path).NewWriter(ctx)
	if _, err := w.Write(fileContents); err != nil {
		logrus.Errorf("createFile: unable to write file %s: %v", path, err)
		return err
	}
	if err := w.Close(); err != nil {
		logrus.Errorf("createFile: unable to close bucket: %v", err)
		return err
	}
	return nil
}

// GetFilesFromStorageBucket gets all files at path
func GetFilesFromStorageBucket(bucketName string, path string) (map[string][]byte, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	bucket := client.Bucket(bucketName)
	// return nil
	files, err := listFilesInBucket(bucket, bucketName, path)
	if err != nil {
		return nil, err
	}
	fileMap := make(map[string][]byte)
	for _, file := range files {
		reader, err := bucket.Object(file).NewReader(ctx)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		contents, err := ioutil.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		fileMap[file] = contents
	}
	return fileMap, err
}

func listFilesInBucket(bucket *storage.BucketHandle, bucketName, path string) ([]string, error) {
	ctx := context.Background()
	query := &storage.Query{Prefix: path}
	if path == "" {
		query = nil
	}
	logrus.Infof("Querying %s for %s", bucketName, path)
	it := bucket.Objects(ctx, query)
	var files []string
	for {
		obj, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			logrus.Errorf("listBucket: unable to list files in bucket at %s, err: %v", path, err)
			return nil, err
		}
		files = append(files, obj.Name)
	}
	return files, nil
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
