#!/bin/bash
set -eo pipefail
ARTIFACT_BUCKET=$(cat bucket-name.txt)
cd function

# The zipped files do not have the currect permissions on Windows
# Get the name of the zip file
output=`aws s3 ls $ARTIFACT_BUCKET`
words=($output)
ARTIFACT_ZIP=${words[3]}

# Zip up HandleServiceEvent executable
build-lambda-zip.exe -o $ARTIFACT_ZIP HandleServiceEvent
# Now push it to the S3 bucket
aws s3 cp ${ARTIFACT_ZIP} s3://${ARTIFACT_BUCKET}/${ARTIFACT_ZIP}
