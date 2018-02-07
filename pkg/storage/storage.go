package storage

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	"log"
	"os"
	"time"

	"golang.org/x/net/context"

	"cloud.google.com/go/iam"
	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

var bucketName string

// CreateStorageBucket creates a storage bucket to store the source context in
func createStorageBucket(sc string) error {
	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	projectID, err := getProjectID(sc)
	bucketName := fmt.Sprintf("kbuild-buckets-%d", time.Now().Unix())
	if err := create(client, projectID, bucketName); err != nil {
		log.Fatal(err)
	}
	logrus.Info("Created bucket %s in project %s", bucketName, projectID)

	if err := client.Bucket(bucketName).Create(ctx, projectID, nil); err != nil {
		return err
	}
	return UploadContextToStorageBucket
}

// GetBucketName returns the bucket name
func GetBucketName() string {
	return bucketName
}

// DeleteStorageBucket deletes the storage bucket the source context is in
func DeleteStorageBucket() error {

	return nil
}

// UploadSourceContextToStorageBucket uploads the source context to the storage bucket
func UploadSourceContextToStorageBucket(sc string) error {
	if err := createStorageBucket(sc); err != nil {
		return err
	}
	ctx := context.Background()
	filepath.Walk(sc, func(path string, info os.FileInfo, err error) error {
		f, err := os.Open(info.Name())
		if err != nil {
			return err
		}
		defer f.Close()

		wc := client.Bucket(bucketName).Object(object).NewWriter(ctx)
		if _, err = io.Copy(wc, f); err != nil {
			return err
		}
		if err := wc.Close(); err != nil {
			return err
		}
		return nil
	})

	return nil
}

// GetFilesFromStorageBucket gets files/dirs that match the context
func GetFilesFromStorageBucket(context string) error {

	return nil
}

func getProjectID(scope string) (string, error) {
	ctx := context.Background()
	defaultCreds, err := google.FindDefaultCredentials(ctx, scope)
	if err != nil {
		return "", err
	}
	return defaultCreds.ProjectID, nil
}
