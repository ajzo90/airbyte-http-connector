echo "building source..."
docker build -t ajzo90/airbyte-source-go-example:0.1.0 -f Dockerfile-source .

echo "building destination..."
docker build -t ajzo90/airbyte-destination-go-example:0.1.0 -f Dockerfile-destination .

echo "building example server..."
docker build -t ajzo90/server:dev -f Dockerfile-server .

#docker build -t airbyte/worker:dev -f Dockerfile-worker .

#go build -o worker ./cmd/worker/