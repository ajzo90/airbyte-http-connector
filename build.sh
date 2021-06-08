SRC=ajzo90/airbyte-source-go-example:0.1.1
DEST=ajzo90/airbyte-destination-go-example:0.1.1

echo "building source...$SRC"
docker build -t $SRC -f Dockerfile-source .
echo "push source...$SRC"

echo "building destination $DEST..."
docker build -t $DEST -f Dockerfile-destination .
echo "push destination...$DEST"
docker push $DEST

echo "building example server..."
docker build -t ajzo90/server:dev -f Dockerfile-server .

#docker build -t airbyte/worker:dev -f Dockerfile-worker .

#go build -o worker ./cmd/worker/