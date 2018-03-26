# dupe2bucket

Copies RDS Log Files to an S3 bucket.

## Usage

Copy all logs from all instances to YourBucket.

    dupe2bucket -bucket=YourBucket

Copy all logs for a specific RDS instance identifier.

    dupe2bucket -bucket=YourBucket -instance=YourRDSInstance

Copy only logs with 'error' in their name.

    dupe2bucket -bucket=YourBucket -filter_fn=error

Copy only logs with 'error' in their name for a specific RDS instance identifier.

    dupe2bucket -bucket=YourBucket -instance=YourRDSInstance -filter_fn=error

Copy all logs from all instances to YourBucket with an object name prefix.

    dupe2bucket -bucket=YourBucket -bucket-prefix="some/prefix"

## Copy Behavior

All logs in scope are copied by default. However, successive runs will only copy logs that have been written to since the last run. State files (see below) limit successive runs to choosing only the logs with more recent writes.

## Bucket Contents

Your bucket will contain log files and state files after a successful run.

### Log Files

{bucket-prefix}/Logs/{db instance arn}/YYYY/ISOWeek/{log file name}

These objects will contain whatever RDS logged to the respective file.

### State Files

{bucket-prefix}/State/{db instance arn}

These objects are state files who, once created, only have their tag values changed. Removing the tags, or the file, will cause dupe2bucket to run a complete copy of all logs in scope.

## Bucket Lifecycle

Lifecycle policies should refrain from making any changes to the state file objects.

## Limitations

The RDS API method [DownloadDBLogFilePortion](https://docs.aws.amazon.com/sdk-for-go/api/service/rds/) is limited to 1 MB per request. Large log files will require multiple RDS API requests.

dupe2bucket uses an in-memory structure to temporarily hold log file data. Copying very large log files may exceed the available RAM of the host executing dupe2bucket. In those cases, the host may begin to use swap space which may be considerably slower than RAM.

Log files are copied one at time given the current model of using RAM.
