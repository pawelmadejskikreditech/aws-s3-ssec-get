# AWS S3 SSE-C GET

CLI tool for downloading encrypted content from S3

## Install

    go get github.com/pawelmadejskikreditech/aws-s3-ssec-get

## Usage

Export environment variables on your machine (required ENVs are defined in `.env.example` file)
You can create your own `.env` file end export it using following command:

    export $(cat .env | xargs)

Following flags are available. `bucket` and `path` are required. `key` and `key-file` are mutual exclusive.
By default tool prints items to `Stdout`.

    aws-s3-ssec-get -help
    Usage of aws-s3-ssec-get:
    -bucket string
            AWS bucket name
    -key string
            base64 encoded encryption key string
    -key-file string
            file with binary encryption key
    -output string
            output file (default stdout)
    -path string
            AWS item path


## Example

    aws-s3-ssec-get -bucket dev-adnetwork -path pawel-kontomierz/f3dba158-3b05-43db-8c66-0c7870d82165/message -key mVgpimdFzFFkRjsqLzSOouDofSIpY8fqbmDIJUcak84= -output message.html
