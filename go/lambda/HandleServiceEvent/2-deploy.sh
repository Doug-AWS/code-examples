#!/bin/bash
set -eo pipefail
ARTIFACT_BUCKET=$(cat bucket-name.txt)
cd function
GOOS=linux go build -o HandleServiceEvent HandleServiceEvent.go
cd ../
aws cloudformation package --template-file template.yml --s3-bucket $ARTIFACT_BUCKET --output-template-file out.yml
aws cloudformation deploy --template-file out.yml --stack-name blank-go --capabilities CAPABILITY_NAMED_IAM

# The zipped files do not have the currect permissions on Windows
# Get the name of the zip file
output=`aws s3 ls $ARTIFACT_BUCKET`
words=($output)
ARTIFACT_ZIP=${words[3]}

cd function

# Zip up HandleServiceEvent executable
build-lambda-zip.exe -o $ARTIFACT_ZIP HandleServiceEvent
# Now push it to the S3 bucket
aws s3 cp ${ARTIFACT_ZIP} s3://${ARTIFACT_BUCKET}/${ARTIFACT_ZIP}

cd ..
