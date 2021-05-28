echo "building source..."
docker build -t airbyte/source-go-example:dev -f Dockerfile-source .

echo "building destination..."
docker build -t airbyte/destination-go-example:dev -f Dockerfile-destination .

echo "building example server..."
docker build -t airbyte/server:dev -f Dockerfile-server .

#docker build -t airbyte/worker:dev -f Dockerfile-worker .

#go build -o worker ./cmd/worker/