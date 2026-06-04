#!/bin/sh
set -eu

BOOTSTRAP_SERVER="${KAFKA_BOOTSTRAP_SERVER:-kafka:9092}"
TOPICS_CSV="${KAFKA_TOPICS:-}"

if [ -z "$TOPICS_CSV" ]; then
  echo "No Kafka topics configured. Skipping initialization."
  exit 0
fi

echo "Waiting for Kafka at ${BOOTSTRAP_SERVER}..."
until /opt/kafka/bin/kafka-topics.sh --bootstrap-server "$BOOTSTRAP_SERVER" --list >/dev/null 2>&1; do
  sleep 2
done

OLD_IFS="$IFS"
IFS=','
for topic in $TOPICS_CSV; do
  topic_trimmed="$(echo "$topic" | xargs)"
  if [ -n "$topic_trimmed" ]; then
    echo "Ensuring topic exists: $topic_trimmed"
    /opt/kafka/bin/kafka-topics.sh \
      --bootstrap-server "$BOOTSTRAP_SERVER" \
      --create \
      --if-not-exists \
      --topic "$topic_trimmed" \
      --partitions 1 \
      --replication-factor 1
  fi
done
IFS="$OLD_IFS"

echo "Kafka topic initialization completed."
