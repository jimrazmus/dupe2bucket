package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// GetListOfLogFiles returns a slice of DescribeDBLogFiles.LogFileName, the greatest LastWritten value, and error
func GetListOfLogFiles(rdsHandle *rds.RDS, dbii string, fn string, lw int64) ([]string, int64, error) {

	var l []string
	var t int64 = 0

	input := &rds.DescribeDBLogFilesInput{
		DBInstanceIdentifier: aws.String(dbii),
		FilenameContains:     aws.String(fn),
		// FileSize:             aws.Int64(fs),
		FileLastWritten: aws.Int64(lw),
	}

	result, err := rdsHandle.DescribeDBLogFiles(input)
	// fmt.Println(result)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rds.ErrCodeDBInstanceNotFoundFault:
				fmt.Println(rds.ErrCodeDBInstanceNotFoundFault, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return l, t, err
	}

	for _, lf := range result.DescribeDBLogFiles {
		// fmt.Println(*lf.LogFileName)
		l = append(l, *lf.LogFileName)
		if *lf.LastWritten > t {
			t = *lf.LastWritten
		}
	}

	return l, t, nil
}

// GetLogFile returns a string of DownloadDBLogFilePortionOutput.LogFileData, error
func GetLogFile(rdsHandle *rds.RDS, dbii string, fn string) (string, error) {

	var sb strings.Builder

	input := &rds.DownloadDBLogFilePortionInput{
		DBInstanceIdentifier: aws.String(dbii),
		LogFileName:          aws.String(fn),
		Marker:               aws.String("0"),
		NumberOfLines:        aws.Int64(0),
	}

	result, err := rdsHandle.DownloadDBLogFilePortion(input)
	// fmt.Println(result)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rds.ErrCodeDBInstanceNotFoundFault:
				fmt.Println(rds.ErrCodeDBInstanceNotFoundFault, aerr.Error())
			case rds.ErrCodeDBLogFileNotFoundFault:
				fmt.Println(rds.ErrCodeDBLogFileNotFoundFault, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return "", err
	}

	_, err = sb.WriteString(*result.LogFileData)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	for *result.AdditionalDataPending == true {
		input := &rds.DownloadDBLogFilePortionInput{
			DBInstanceIdentifier: aws.String(dbii),
			LogFileName:          aws.String(fn),
			Marker:               aws.String(*result.Marker),
			NumberOfLines:        aws.Int64(0),
		}
		result, _ := rdsHandle.DownloadDBLogFilePortion(input)
		_, err = sb.WriteString(*result.LogFileData)
		if err != nil {
			fmt.Println(err.Error())
			return "", err
		}
	}

	return sb.String(), nil
}

// GetMapOfInstances returns a map of [DBInstanceIdentifier]DBInstanceArn, error
func GetMapOfInstances(rdsHandle *rds.RDS, dbii string) (map[string]string, error) {

	m := make(map[string]string)

	input := &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(dbii),
	}

	result, err := rdsHandle.DescribeDBInstances(input)
	// fmt.Println(result)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rds.ErrCodeDBInstanceNotFoundFault:
				fmt.Println(rds.ErrCodeDBInstanceNotFoundFault, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return m, err
	}

	for _, db := range result.DBInstances {
		// fmt.Println(*db.DBInstanceArn, "--", *db.DBInstanceIdentifier)
		m[*db.DBInstanceIdentifier] = *db.DBInstanceArn
	}

	return m, nil
}

// S3upload conveniently wraps putting S3 objects and returns error
func S3upload(s3Handle *s3manager.Uploader, bucket string, key string, body string) error {

	input := &s3manager.UploadInput{
		Body:                 aws.ReadSeekCloser(strings.NewReader(body)),
		Bucket:               aws.String(bucket),
		Key:                  aws.String(key),
		ServerSideEncryption: aws.String("AES256"),
	}

	_, err := s3Handle.Upload(input)
	// fmt.Println(result)

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	return nil
}

// S3GetState attempts to fetch the "lastWrite" tag and return its value
func S3GetState(s3Handle *s3.S3, bucket string, key string) (int64, error) {

	input := &s3.GetObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result, err := s3Handle.GetObjectTagging(input)
	// fmt.Println(result)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				// fmt.Println(key, s3.ErrCodeNoSuchKey, aerr.Error())
				return -1, nil
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return 0, err
	}

	for _, tag := range result.TagSet {
		// fmt.Println(*tag.Key, *tag.Value)
		if *tag.Key == "lastWrite" {
			i, err := strconv.ParseInt(*tag.Value, 10, 64)
			if err != nil {
				fmt.Println(err)
				return 0, err
			}
			return i, nil
		}
	}

	return 0, nil
}

// S3SetState attempts to add/update the "lastWrite" tag.
func S3SetState(s3Handle *s3.S3, bucket string, key string, lastWrite int64) (int64, error) {

	lw := strconv.FormatInt(lastWrite, 10)

	input := &s3.PutObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Tagging: &s3.Tagging{
			TagSet: []*s3.Tag{
				{
					Key:   aws.String("lastWrite"),
					Value: aws.String(lw),
				},
			},
		},
	}

	_, err := s3Handle.PutObjectTagging(input)
	// fmt.Println(result)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				fmt.Println(s3.ErrCodeNoSuchKey, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return 0, err
	}

	return lastWrite, nil
}

func main() {
	var (
		bucket        = flag.String("bucket", "", "S3 Bucket to put log files")
		bucket_prefix = flag.String("bucket-prefix", "", "Optional S3 Bucket prefix for the log files.")
		instance      = flag.String("instance", "", "Optionally target a specific RDS instance name. e.g. -instance=my-database")
		filter_fn     = flag.String("filter-fn", "", "Optional string to filter RDS log file names. e.g. -filter-fn=error")
	)
	flag.Parse()

	if *bucket == "" {
		fmt.Print("\n*** A bucket must be specified ***\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	aws_session := session.Must(session.NewSession())
	rds_client := rds.New(aws_session)
	uploader := s3manager.NewUploader(aws_session)
	s3_client := s3.New(aws_session)

	year_t, week_t := time.Now().UTC().ISOWeek()
	year := strconv.Itoa(year_t)
	week := strconv.Itoa(week_t)

	update_state := true

	// retrieve candidate RDS instances
	m, err := GetMapOfInstances(rds_client, *instance)
	if err != nil {
		os.Exit(1)
	}

	// iterate over the RDS instances
	for dbname, dbarn := range m {

		fmt.Println("Copying logs for:", dbarn)

		// retrieve lastWrite state for the dbarn
		s := []string{*bucket_prefix, "LastWrittenState", dbarn}
		state_key := strings.Join(s, "/")
		lastWrite, err := S3GetState(s3_client, *bucket, state_key)
		if err != nil {
			break
		}

		if lastWrite == -1 {
			// Create the state object
			err = S3upload(uploader, *bucket, state_key, "DO NOT DELETE")
			if err != nil {
				break
			}
			lastWrite = 0
		}

		logs, max_lw, err := GetListOfLogFiles(rds_client, dbname, *filter_fn, lastWrite)
		if err != nil {
			break
		}

		// iterate over the list of logs
		for _, logname := range logs {
			fmt.Println("\t", logname)

			logdata, err := GetLogFile(rds_client, dbname, logname)
			if err != nil {
				update_state = false
				break
			}

			s = []string{*bucket_prefix, dbarn, year, week, logname}
			object_key := strings.Join(s, "/")
			err = S3upload(uploader, *bucket, object_key, logdata)
			if err != nil {
				update_state = false
				break
			}
		}

		if update_state {
			_, _ = S3SetState(s3_client, *bucket, state_key, max_lw)
		}
	}
}
