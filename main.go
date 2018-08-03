package main

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
)

type writerAtCloser interface {
	io.WriterAt
	io.Closer
}

type stdout struct {
	f *os.File
}

func newStdout() (writerAtCloser, error) {
	tempFile, err := ioutil.TempFile("", "aws-s3-ssec-get")
	if err != nil {
		return nil, err
	}
	return &stdout{f: tempFile}, nil
}

func (s *stdout) WriteAt(p []byte, off int64) (n int, err error) {
	return s.f.WriteAt(p, off)
}

func (s *stdout) Close() error {
	defer os.Remove(s.f.Name())
	defer s.f.Close()

	buffer := bytes.NewBuffer([]byte{})
	buffer.ReadFrom(s.f)

	fmt.Fprint(os.Stdout, buffer.String())
	return nil
}

func main() {
	path := flag.String("path", "", "AWS item path")
	bucket := flag.String("bucket", "", "AWS bucket name")
	key := flag.String("key", "", "base64 encoded encryption key string")
	keyFile := flag.String("key-file", "", "file with binary encryption key")
	output := flag.String("output", "", "output file (default stdout)")

	flag.Parse()

	if path == nil || *path == "" {
		log.Fatal("path arg is required")
	}
	if bucket == nil || *bucket == "" {
		log.Fatal("bucket arg is required")
	}

	encryptionKey, err := decodeEncryptionKey(key, keyFile)
	if err != nil {
		log.Fatal(err)
	}

	out, err := defineOutput(output)
	if err != nil {
		log.Fatalf("unable to open output: %v", err)
	}
	defer out.Close()

	s := session.Must(session.NewSessionWithOptions(session.Options{SharedConfigState: session.SharedConfigEnable}))
	if err != nil {
		log.Fatalf("unable to create AWS session: %v", err)
	}

	downloader := s3manager.NewDownloader(s)
	_, err = downloader.Download(out, &s3.GetObjectInput{
		Bucket:               bucket,
		Key:                  path,
		SSECustomerAlgorithm: aws.String(s3.ServerSideEncryptionAes256),
		SSECustomerKey:       aws.String(string(encryptionKey)),
		SSECustomerKeyMD5:    aws.String(md5sum(encryptionKey)),
	})
	if err != nil {
		log.Fatalf("unable to download item from S3: %v", err)
	}
}

func decodeEncryptionKey(key, file *string) ([]byte, error) {
	var encryptionKey []byte
	var err error

	if key != nil && *key != "" {
		encryptionKey, err = base64.StdEncoding.DecodeString(*key)
		if err != nil {
			return nil, errors.Wrap(err, "unable to decode encryption key: %v")
		}
	} else if file != nil && *file != "" {
		encryptionKey, err = ioutil.ReadFile(*file)
		if err != nil {
			return nil, errors.Wrap(err, "unable to read encryption key file %s: %v")
		}
	}
	return encryptionKey, nil
}

func defineOutput(loc *string) (writerAtCloser, error) {
	if loc == nil || *loc == "" {
		return newStdout()
	}

	return os.Create(*loc)
}

func md5sum(encryptionKey []byte) string {
	sum := md5.Sum(encryptionKey)
	sum64 := make([]byte, base64.StdEncoding.EncodedLen(len(sum)))
	base64.StdEncoding.Encode(sum64, sum[:])
	return string(sum64)
}
