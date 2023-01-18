#!/bin/bash

~/kafka_2.13-3.0.0/bin/kafka-topics.sh --create --topic ya2k-clicks --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1
~/kafka_2.13-3.0.0/bin/kafka-topics.sh --create --topic ya2k-tinurls --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1
