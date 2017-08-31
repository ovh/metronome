FROM golang:1.7.4
MAINTAINER d33d33 <kevin@d33d33.fr>

EXPOSE 8080

# Setup work directory
RUN mkdir -p /go/src/github.com/ovh/metronome
WORKDIR /go/src/github.com/ovh/metronome

# Get wait-for-it
ADD https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh ./wait-for-it.sh
RUN chmod +x ./wait-for-it.sh

# Install glide
RUN curl https://glide.sh/get | sh

# Setup GO ENV
ENV GOPATH /go
ENV PATH $PATH:/usr/local/go/bin:$GOPATH/bin

# Install dependencies
COPY glide.yaml ./glide.yaml
COPY glide.lock ./glide.lock
RUN glide install

# Build metronome
COPY Makefile ./Makefile
COPY src ./src
RUN make
RUN make install

# Use default config
COPY default.yaml ./default.yaml
