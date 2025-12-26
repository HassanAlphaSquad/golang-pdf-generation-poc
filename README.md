# PDF Generation API

Simple POC for generating PDFs from HTML using headless Chrome.

## What it does

Takes HTML and converts it to PDF files. That's it.

## Running it

```bash
make run
```

Server starts on port 3000. Swagger docs available at http://localhost:3000/swagger/index.html

## Usage

### Generate PDF from HTML

```bash
curl -X POST http://localhost:3000/api/v1/pdf/generate \
  -H "Content-Type: application/json" \
  -d '{
    "html": "<html><body><h1>Test</h1></body></html>",
    "filename": "test.pdf"
  }'
```

Returns a job ID. Use it to check status:

```bash
curl http://localhost:3000/api/v1/pdf/status/{job_id}
```

Download when complete:

```bash
curl http://localhost:3000/api/v1/pdf/download/{job_id} -o output.pdf
```

### Generate from URL

```bash
curl -X POST http://localhost:3000/api/v1/pdf/generate/url \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com",
    "filename": "webpage.pdf"
  }'
```

## Available commands

```bash
make run      # Generate swagger and start server
```

## Config

Set via environment variables:
- `PORT` - Server port (default: 3000)
- `OUTPUT_DIR` - Where PDFs are saved (default: ./output)

## How it works

1. You send HTML
2. Job gets queued
3. Headless Chrome renders HTML
4. PDF gets generated
5. You download it

Jobs are cleaned up after 24 hours.

## Requirements

- Go 1.21+
- Chrome/Chromium (for headless rendering)
