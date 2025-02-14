docker run -p 9001:9001 \
  -e PORT=9001 \
  -e MAX_CONN=2 \
  -e POW_DIFFICULTY=20 \
  wise-tcp-server
#docker run -p 9001:9001 --env-file docker/server.env wise-tcp-server
