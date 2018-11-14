package main

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"golang.org/x/sync/errgroup"
)

var (
	downloader *s3manager.Downloader
	uploader   *s3manager.Uploader
	s3Conn     *s3.S3
)

//init - initializes sdk client connections
func init() {
	s := session.Must(session.NewSession())
	s3Conn = s3.New(s)
	downloader = s3manager.NewDownloader(s)
	uploader = s3manager.NewUploader(s)
}

//main - boilerplate code for aws lambda
func main() {
	lambda.Start(Handler)
}

//Handler - The main lambda handler
func Handler(ctx context.Context, s3Event events.S3Event) error {
	for _, record := range s3Event.Records {
		key := record.S3.Object.Key
		bucket := record.S3.Bucket.Name
		log.Printf("Key: %s, Bucket: %s\n", key, bucket)

		if !isZipFile(key) {
			log.Printf("Skipping non-zip file %s\n", key)
			continue
		}
		//need the prefix to know exactly where to send this
		prefix := getPrefixForS3Key(key)

		path, err := createTempDirectory(prefix)
		if err != nil {
			log.Fatal(err)
			return err
		}

		downloadedZipPath, err := download(bucket, key, path)
		if err != nil {
			log.Fatal(err)
			return err
		}

		if err := unzip(downloadedZipPath, path); err != nil {
			log.Fatal(err)
			return err
		}

		dstBucket := os.Getenv("DST_BUCKET")
		if err := uploadAll(path, dstBucket); err != nil {
			log.Fatal(err)
			return err
		}

		//clean up
		deleteObject(key, bucket)
		return nil
	}
	return nil
}

func getPrefixForS3Key(key string) string {
	prefix := filepath.Dir(key)
	if prefix == "." {
		prefix = ""
	} else {
		prefix = "/" + prefix
	}
	return prefix
}

//isZipFile is a helper function to determine if a file is a zip file or not
func isZipFile(fileName string) bool {
	extension := filepath.Ext(fileName)
	return extension == ".zip"
}

//createTempDiectory initializes a temp directory
func createTempDirectory(prefix string) (string, error) {
	const tempPath = "/tmp"
	const dirPerm = 0777
	now := strconv.Itoa(int(time.Now().UnixNano()))
	path := fmt.Sprintf("%s/%s%s", tempPath, now, prefix)
	if _, err := os.Stat(path); err == nil {
		if err := os.RemoveAll(path); err != nil {
			return "", err
		}
	}

	if err := os.MkdirAll(path, dirPerm); err != nil {
		return "", err
	}
	log.Printf("Created temp dir: %s", path)
	return path, nil
}

//download - downloads a object from s3 into the given directory using the key as the base filename
func download(bucket, key, path string) (string, error) {
	fileName := path + "/" + filepath.Base(key)
	file, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})

	if err != nil {
		return "", err
	}
	log.Println("Downloaded", fileName, numBytes, "bytes")

	return fileName, nil
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			f, err := os.OpenFile(
				path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
			log.Printf("Inflating file: %s\n", f.Name())
		}
	}
	//delete the original zip file - dont want to reupload that
	os.Remove(src)
	return nil
}

//upload - uploads a single file to an s3 bucket
func upload(fileName, path, bucket string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	elements := strings.Split(path, "/")
	//log.Printf("len: %d, elements: %v", len(elements), elements)
	prefix := strings.Join(elements[:3], "/") + "/"
	//log.Printf("file.Name(): %s prefix: %s\n", file.Name(), prefix)
	key := strings.Replace(file.Name(), prefix, "", 1)
	log.Printf("key: %s, bucket: %s", key, bucket)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return err
	}
	return nil
}

//uploadAll - Helper method to upload all files in a directory path to an s3 bucket
func uploadAll(path, bucket string) error {
	eg := errgroup.Group{}

	err := filepath.Walk(path, func(fileName string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println(err)
			return err
		}
		if info.IsDir() {
			return nil
		}
		eg.Go(func() error {
			return upload(fileName, path, bucket)
		})
		return nil
	})

	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func deleteObject(key, bucket string) error {
	log.Printf("Deleting original s3 object (%s::%s)", bucket, key)
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	_, err := s3Conn.DeleteObject(input)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}
