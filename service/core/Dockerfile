FROM golang:1.16 as gobuild

RUN apk update && apk upgrade && apk add --no-cache ca-certificates && update-ca-certificates 2>/dev/null || true

ARG ACCESS_TOKEN_USR="nothing"
ARG ACCESS_TOKEN_PWD="nothing"


RUN printf "machine github.com\n\
    login ${ACCESS_TOKEN_USR}\n\
    password ${ACCESS_TOKEN_PWD}\n\
    \n\
    machine api.github.com\n\
    login ${ACCESS_TOKEN_USR}\n\
    password ${ACCESS_TOKEN_PWD}\n"\
    >> /root/.netrc
RUN chmod 600 /root/.netrc

WORKDIR ${GOPATH}/src/services/
ENV GO111MODULE=on

# Git is required for fetching the dependencies.
RUN apt-get update && apt-get install -y git && rm -rf /var/lib/apt/lists/*

# Copy only go.mod/go.sum to cache dependencies between local docker builds
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

# Always copy all source codes except the ones in .dockerignore
COPY . .

# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /core/coreservice . && chmod 0755 /core/coreservice
COPY --chmod=0644 --chown=root:root ./config.json /core/
COPY --chmod=0644 --chown=root:root ./templates /core/templates

############################
# STEP 2 build a small image
############################
FROM scratch

# copy the ca-certificate.crt from the build stage
COPY --from=gobuild /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Copy our static executable.
COPY --from=gobuild /core /

EXPOSE 3000

ENTRYPOINT ["/coreservice"]
