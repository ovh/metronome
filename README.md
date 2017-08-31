
<h1><img src="https://rawgithub.com/ovh/metronome/master/icon.svg" width="48" height="48">&nbsp;Metronome - Distributed, fault tolerant scheduler</h1>
[![version](https://img.shields.io/badge/status-alpha-orange.svg)](https://github.com/ovh/metronome)
[![Build Status](https://travis-ci.org/ovh/metronome.svg?branch=ci)](https://travis-ci.org/ovh/metronome)
[![codecov](https://codecov.io/gh/ovh/metronome/branch/master/graph/badge.svg)](https://codecov.io/gh/ovh/metronome)
[![Go Report Card](https://goreportcard.com/badge/github.com/ovh/metronome)](https://goreportcard.com/report/github.com/ovh/metronome)
[![GoDoc](https://godoc.org/github.com/ovh/metronome?status.svg)](https://godoc.org/github.com/ovh/metronome)
[![Join the chat at https://gitter.im/ovh-metronome/Lobby](https://badges.gitter.im/ovh-metronome/Lobby.svg)](https://gitter.im/ovh-metronome/Lobby?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

Metronome is a distributed and fault-tolerant event scheduler. It can be used to trigger remote systems throught events (HTTP, AMQP, KAFKA).

Metronome is written in Go and leverage the power of kafka streams to provide fault tolerance, reliability and scalability.

Metronome take a new approach to the job scheduling problem, as it only handle job scheduling not effective execution. Effective job execution is perform by triggered external system.

Metronome has a number of advantages over regular cron:
- Jobs can be written in any language, using any technology as it only trigger event.
- Jobs are schedule using [ISO8601][ISO8601] repeating interval notation, which enables more flexibility.
- It is able to handle high volumes of scheduled jobs in a completely fault way.
- Easy admin, thanks to a great [UI][UI].

## Status

Currently Metronome is under heavy development.

## Quick start

The best way to test and develop Metronome is using docker, you will need [Docker Toolbox](https://www.docker.com/docker-toolbox) installed before proceding.

- Clone the repository.

- Run the included Docker Compose config:

`docker-compose up -d`

This will start, PostgreSQL, Redis, Kafka and Metronome instances.

Open your browser and navigate to `localhost:8081`

## Contributing

Instructions on how to contribute to Metronome are available on the [Contributing][Contributing] page.

## Get in touch

- Twitter: [@notd33d33](https://twitter.com/notd33d33)

[UI]: https://github.com/ovh/metronome-ui
[Contributing]: CONTRIBUTING.md
[ISO8601]: http://en.wikipedia.org/wiki/ISO_8601 "ISO8601 Standard"
