package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	// "github.com/aws/aws-sdk-go/service/s3"

	"github.com/x1um1n/checkerr"
)

var sess *session.Session

func connectAWS() *session.Session {
	sess, err := session.NewSession(
		&aws.Config{
			Region:      aws.String(K.String("s3_region")),
			Credentials: credentials.NewStaticCredentials(K.String("s3_access_key_id"), K.String("s3_secret_access_key"), ""),
		},
	)
	checkerr.Check(err, "Error connecting to S3")

	return sess
}
