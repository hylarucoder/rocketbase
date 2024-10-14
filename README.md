<p align="center">
    <a href="https://pocketbase.io" target="_blank" rel="noopener">
        <img src="https://i.imgur.com/5qimnm5.png" alt="RocketBase - open source backend in 1 file" />
    </a>
</p>

## Rocketbase

> Rocketbase = Pocketbase + Postgres + a lot of goodies

## Credit

- pocketbase for main codebase
- postgresbase for adapting postgres

## Usage

You can easily fork and setup the project.

```bash
# clone and download libraries
git clone https://github.com/AlperRehaYAZGAN/postgresbase
cd postgresbase
go mod download

# docker-compose has 3 service for test pocketbase all features:
# 1. Postgres: runs on port 5432
# 1. postgres://user:pass@localhost/logs?sslmode=disable
# 2. minio: UI runs on port 9001 and API on 9000  (minio123:minio123)
# 2. s3://minio123:minio123@localhost:9000/public
# (dont forget to manually create bucket called "public" via web ui to establish s3 connection from pocketbase)
# 3. mailhog: port: SMTP-1025 and UI-8025
# 3. smtp://localhost:1025 - http://localhost:8025
docker-compose up -d

# before run the project, you need to create and set RSA Public key pair for JWT before run the application.
# you can use following command to generate RSA key pair
openssl genrsa -out ./keys/private.pem 2048
openssl rsa -in ./keys/private.pem -outform PEM -pubout -out ./keys/public.pem

# after generating keys, you can set as environment variables
export JWT_PRIVATE_KEY=$(cat ./keys/private.pem)
export JWT_PUBLIC_KEY=$(cat ./keys/public.pem)
export CGO_ENABLED=0
export LOGS_DATABASE="postgresql://user:pass@localhost/logs?sslmode=disable"
export DATABASE="postgresql://user:pass@localhost/postgres?sslmode=disable"

# optional ENV_VARS
export BCRYPT_COST=10 # default is 12

# export is success you can run the project ✅
go run -tags pq ./examples/base serve  

```
