#!/usr/bin/env bash

docker run -p 9090:9090 -v $PWD/config/prometheus.yml:/etc/prometheus/prometheus.yml --network host prom/prometheus
