#!/bin/sh

rm upcycling-xmas-tree.tar upcycling-xmas-tree.shi

# 1. Generate arm executable for upcycling-xmas-tree-service server
cd service
./make-cross-build.sh

# 2. Make one pack file for upcycling-xmas-tree docker image
cd ../local_files
./pack-files.sh

# 3. Make upcycling-xmas-tree docker image
cd ..
docker build --rm -t upcycling-xmas-tree .
docker save upcycling-xmas-tree > upcycling-xmas-tree.tar
zip upcycling-xmas-tree.shi upcycling-xmas-tree.tar upcycling-xmas-tree.sh
