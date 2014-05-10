package upload

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/meatballhat/artifacts/path"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

type uploader struct {
	BucketName string
	Paths      *path.PathSet
	TargetPath string
}

// Upload does the deed!
func Upload(opts *Options) {
	newUploader(opts).Upload()
}

func newUploader(opts *Options) *uploader {
	u := &uploader{
		BucketName: opts.BucketName,
		TargetPath: opts.TargetPath,
		Paths:      path.NewPathSet(),
	}

	for _, s := range strings.Split(opts.Paths, ";") {
		trimmed := strings.TrimSpace(s)
		if len(trimmed) > 0 {
			u.Paths.Add(path.NewPath(opts.WorkingDir, trimmed, ""))
		}
	}

	return u
}

func (u *uploader) Upload() error {
	auth, err := aws.GetAuth("", "")
	if err != nil {
		return err
	}

	conn := s3.New(auth, aws.USEast)
	bucket := conn.Bucket(u.BucketName)

	if bucket == nil {
		return fmt.Errorf("failed to get bucket")
	}

	for artifact := range u.files() {
		u.uploadFile(artifact)
	}

	return nil
}

func (u *uploader) files() chan *artifact {
	artifacts := make(chan *artifact)

	go func() {
		for _, path := range u.Paths.All() {
			to, from, root := path.To, path.From, path.Root
			if path.IsDir() {
				root = filepath.Join(root, from)
				if strings.HasSuffix(root, "/") {
					root = root + "/"
				}
			}

			filepath.Walk(path.Fullpath(), func(f string, info os.FileInfo, err error) error {
				if info.IsDir() {
					return nil
				}

				relPath := strings.Replace(strings.Replace(f, root, "", -1), root+"/", "", -1)
				destination := relPath
				if len(to) > 0 {
					if path.IsDir() {
						destination = filepath.Join(to, relPath)
					} else {
						destination = to
					}
				}

				artifacts <- &artifact{Source: f, Destination: destination}
				return nil
			})

		}
		close(artifacts)
	}()

	return artifacts
}

func (u *uploader) uploadFile(a *artifact) {
	fmt.Printf("Not really uploading %q -> %q\n", a.Source, a.Destination)
}