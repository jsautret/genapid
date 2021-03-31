# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from the latest golang base image
FROM golang:latest as builder

LABEL maintainer="Jérôme Sautret <jerome@sautret.org>"

WORKDIR /app

COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod
# and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory
# inside the container
COPY . .

WORKDIR /app/cmd/genapid

# Build the Go app & Create a default conf to be able to run the
# container without the /conf volume mounted
RUN CGO_ENABLED=0 GOOS=linux go build && \
    mkdir /conf && echo '- name: Test Docker\n\
  pipe:\n\
  - name: Log success Docker\n\
    log:\n\
      msg: Docker Success\n\
' > /conf/api.yml


######## Start a new stage from scratch #######
FROM scratch

# Get the default conf
COPY --from=builder /conf /conf

VOLUME ["/conf"]

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/cmd/genapid .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./genapid", "-port", "8080", "-config", "/conf/api.yml", "-loglevel", "info"]
