FROM golang:1.12-alpine AS builder

# Create the user and group files that will be used in the running container to
# run the process as an unprivileged user.
RUN mkdir /user && \
    echo 'nobody:x:65534:65534:nobody:/:' > /user/passwd && \
    echo 'nobody:x:65534:' > /user/group

# Outside GOPATH since we're using modules.
WORKDIR /src

# Required for fetching dependencies.
RUN apk add --update --no-cache ca-certificates git nodejs nodejs-npm

# Fetch dependencies to cache.
COPY go.mod go.sum ./
# RUN go mod verify

# Add google-cloud json
COPY google-cloud.json /google-cloud.json

# Copy project source files.
COPY . .

# Build static web project.
RUN cd web && npm install && npm run build

# Build.
RUN CGO_ENABLED=0 GOOS=linux go build -installsuffix 'static' -v -o /app .

# Final release image.
FROM alfg/ffmpeg:latest

RUN apk add --update --no-cache ca-certificates python \
        py-pip \
        py-cffi \
        py-cryptography \
      && pip install --upgrade pip \
      && apk add --virtual build-deps \
        gcc \
        libffi-dev \
        python-dev \
        linux-headers \
        musl-dev \
        openssl-dev \
      && pip install gsutil \
      && apk del build-deps \
      && rm -rf /var/cache/apk/*

# Import the user and group files from the first stage.
COPY --from=builder /user/group /user/passwd /etc/

# Import the Certificate-Authority certificates for enabling HTTPS.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Import the project web & executable.
COPY --from=builder /src/web/dist /web/dist
COPY --from=builder /app /app
COPY --from=builder /src/google-cloud.json /google-cloud.json
COPY --from=builder /src/config/dev.yaml /config/default.yml

COPY _boto /root/.boto
COPY _boto /.boto
RUN export BOTO_CONFIG=/root/.boto
RUN export GOOGLE_APPLICATION_CREDENTIALS=/google-cloud.json

EXPOSE 8080

# Perform any further action as an unpriviledged user.
#USER nobody:nobody

# Run binary.
ENTRYPOINT ["/app"]