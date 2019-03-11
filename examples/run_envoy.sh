cd examples
docker build -t envoy:v1 . && \
  docker run \
    --rm \
    -p 9901:9901 -p 10000:10000 \
    -v $PWD/envoy.yaml:/etc/envoy.yaml \
    -it \
    envoy:v1