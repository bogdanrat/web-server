FROM golang:1.16 as gobuild

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
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /auth/authservice . && chmod 0755 /auth/authservice
COPY --chmod=0644 --chown=root:root ./config.json /auth/

############################
# STEP 2 build a small image
############################
FROM scratch
# Copy our static executable.

COPY --from=gobuild /auth /

# used by OpenCensus
EXPOSE 8081
# used by the gRPC Server
EXPOSE 50051

ENTRYPOINT ["/authservice"]
