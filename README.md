# dupe2bucket

Copies RDS Log Files to an S3 bucket.

## Usage

Copy all logs from all instances to YourBucket. Note that successive runs will only copy logs that have been written to since the last run.

    dupe2bucket -bucket=YourBucket

Copy all logs for a specific RDS instance identifier.

    dupe2bucket -bucket=YourBucket -instance=YourRDSInstance

Copy only logs with 'error' in their name.

    dupe2bucket -bucket=YourBucket -filter_fn=error

Copy only logs with 'error' in their name for a specific RDS instance identifier.

    dupe2bucket -bucket=YourBucket -instance=YourRDSInstance -filter_fn=error

Copy all logs from all instances to YourBucket with an object name prefix. Note that successive runs will only copy logs that have been written to since the last run.

    dupe2bucket -bucket=YourBucket -bucket-prefix="some/prefix"

## Bucket Contents

Your bucket will contain log files and state files after a successful run.

### Log Files

{bucket-prefix}/Logs/{db instance arn}/YYYY/ISOWeek/{log file name}

These objects will contain whatever RDS logged to the respective file.

### State Files

{bucket-prefix}/State/{db instance arn}

Note that state files, once created, only have their tag values changed. The values limit successive copies, choosing only the logs with more recent writes. Removing the tags, or the file, will cause dupe2bucket to run a complete copy of all logs in scope.

## Bucket Lifecycle

Lifecycle policies should refrain from making any changes to the state file objects.
