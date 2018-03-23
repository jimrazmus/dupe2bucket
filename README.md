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

Copy all logs from all instances to YourBucket with a object name prefix. Note that successive runs will only copy logs that have been written to since the last run.

    dupe2bucket -bucket=YourBucket -bucket-prefix="some/prefix"
