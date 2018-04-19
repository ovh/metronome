FROM golang:1.10.1
MAINTAINER d33d33 <kevin@d33d33.fr>

EXPOSE 8080

# Setup work directory
RUN mkdir -p /go/src/github.com/ovh/metronome
WORKDIR /go/src/github.com/ovh/metronome

# Get wait-for-it
ADD https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh ./wait-for-it.sh
RUN chmod +x ./wait-for-it.sh

# Install dep
RUN go get -u github.com/golang/dep/...
RUN go get -u github.com/gobuffalo/packr/...

# Setup GO ENV
ENV GOPATH /go
ENV PATH $PATH:/usr/local/go/bin:$GOPATH/bin

# Copy source
COPY src ./src

# Install dependencies
COPY Gopkg.toml ./Gopkg.toml
COPY Gopkg.lock ./Gopkg.lock
RUN dep ensure -v

# Build metronome
COPY Makefile ./Makefile
RUN make install

# Use default config
COPY default.yaml ./default.yaml
