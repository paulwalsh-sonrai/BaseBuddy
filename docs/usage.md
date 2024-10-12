
# BaseBuddy Example Usage

This document provides an example of how to use the BaseBuddy command-line tool to generate documentation from code files.

## Prerequisites

- Go installed on your machine.
- Docker installed (optional for running in a container).
- An AWS S3 bucket configured and the AWS credentials set up on your machine.
- An OpenAI ChatGPT API key.

## Usage

1. **Create a Prompt Template File**

Create a file named `prompt.txt` with the following content:

```
You are a software documenter. Provide a short but informative docs on the code below.
{code}
```

2. **Crawl the Current Directory**

Make sure you have code files in the current directory or subdirectories that you want to document.

3. **Set Up Environment Variables**

Create a `.env` file in the root directory of your project with the following content:

```
S3_BUCKET=your-s3-bucket-name
CHATGPT_API_KEY=your-chatgpt-api-key
```

4. **Run the BaseBuddy Command**

In the root of your project, run the following command:

```bash
go run cmd/basebuddy/main.go gen --prompt prompt.txt
```

Alternatively, if you want to run it within a Docker container:

```bash
docker-compose run basebuddy gen --prompt prompt.txt
```

## Output

After running the command, you should see the generated documentation stored in your specified S3 bucket. Each file's documentation will be saved with its relative path in the bucket, using the `.md` extension.
