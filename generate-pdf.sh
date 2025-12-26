#!/bin/bash

if [ $# -lt 1 ]; then
    echo "Usage: ./generate-pdf.sh <html-file> [output-name.pdf]"
    exit 1
fi

HTML_FILE=$1
OUTPUT_NAME=${2:-"output.pdf"}

if [ ! -f "$HTML_FILE" ]; then
    echo "Error: $HTML_FILE not found"
    exit 1
fi

echo "Reading $HTML_FILE..."
HTML_CONTENT=$(cat "$HTML_FILE" | jq -Rs .)

echo "Sending to API..."
RESPONSE=$(curl -s -X POST http://localhost:3000/api/v1/pdf/generate \
  -H "Content-Type: application/json" \
  -d "{
    \"html\": $HTML_CONTENT,
    \"filename\": \"$OUTPUT_NAME\"
  }")

JOB_ID=$(echo $RESPONSE | jq -r '.job_id')
echo "Job ID: $JOB_ID"

echo "Waiting for PDF generation..."
while true; do
    STATUS_RESPONSE=$(curl -s http://localhost:3000/api/v1/pdf/status/$JOB_ID)
    STATUS=$(echo $STATUS_RESPONSE | jq -r '.status')

    if [ "$STATUS" = "completed" ]; then
        echo "PDF generated successfully!"
        curl -s http://localhost:3000/api/v1/pdf/download/$JOB_ID -o "$OUTPUT_NAME"
        echo "Saved to: $OUTPUT_NAME"
        break
    elif [ "$STATUS" = "failed" ]; then
        ERROR=$(echo $STATUS_RESPONSE | jq -r '.error_message')
        echo "Failed: $ERROR"
        exit 1
    else
        echo "Status: $STATUS"
        sleep 1
    fi
done
