FROM golang:1.13.0-stretch AS builder

# Add Maintainer Info
LABEL maintainer="Vasyl Vaskul <vasyl.vaskul@gmail.com>"

ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    PKG=main

# BUILD
ARG BUILD_TIME
ARG COMMIT_HASH

WORKDIR /build

# Let's cache modules retrieval - those don't change so often
COPY go.mod .
COPY go.sum .
#RUN go mod download

# Copy the code necessary to build the application
# You may want to change this to copy only what you actually need.
COPY . .

RUN echo "BUILD TIME: $BUILD_TIME"
RUN echo "COMMIT HASH: $COMMIT_HASH"

# Set build flags
ENV FLAGS="-X main.BuildTime=$BUILD_TIME -X main.CommitHash=$COMMIT_HASH"

# Build the application
RUN go build -v -ldflags "$FLAGS" -mod vendor -o goskel

# Let's create a /dist folder containing just the files necessary for runtime.
# Later, it will be copied as the / (root) of the output image.
WORKDIR /dist
RUN cp /build/goskel ./goskel

# Optional: in case your application uses dynamic linking (often the case with CGO),
# this will collect dependent libraries so they're later copied to the final image
# NOTE: make sure you honor the license terms of the libraries you copy and distribute
RUN ldd goskel | tr -s '[:blank:]' '\n' | grep '^/' | \
    xargs -I % sh -c 'mkdir -p $(dirname ./%); cp % ./%;'
RUN mkdir -p lib64 && cp /lib64/ld-linux-x86-64.so.2 lib64/

# Copy or create other directories/files your app needs during runtime.
# E.g. this example uses /data as a working directory that would probably
#      be bound to a perstistent dir when running the container normally
RUN mkdir /data

# Create the minimal runtime image
FROM scratch

COPY --chown=0:0 --from=builder /dist /

# Set up the app to run as a non-root user inside the /data folder
# User ID 65534 is usually user 'nobody'.
# The executor of this image should still specify a user during setup.
COPY --chown=65534:0 --from=builder /data /data

# Copy ssl certificates to properly establish an SSL connection
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER 65534
WORKDIR /data

ENTRYPOINT ["/goskel"]
