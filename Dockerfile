# syntax=docker/dockerfile:1

FROM golang:1.22-alpine

# Set destination for COPY
WORKDIR /app

RUN go mod init feedme
RUN go mod tidy

# Download Go modules
#ADD go.sum go.mod  ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/reference/dockerfile/#copy
COPY *.go ./

# Build
RUN go build -o /go-pjlink-emulator

# Optional:
# To bind to a TCP port, runtime parameters must be supplied to the docker command.
# But we can document in the Dockerfile what ports
# the application is going to listen on by default.
# https://docs.docker.com/reference/dockerfile/#expose
EXPOSE 4352

CMD [ "/go-pjlink-emulator" ]