#!/system/bin/sh

TAG_VER=0.3
# -d : Daemon mode
docker run --privileged --net=host -v /dev:/dev -v /data/:/data/ -d upcycling-xmas-tree:$TAG_VER
# -it : interactive mode
#docker run --privileged --net=host -v /dev:/dev -v /data/:/data/ -it upcycling-xmas-tree /usr/bin/start-upcycling-xmas-tree
