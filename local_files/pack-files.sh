#!/bin/sh

rm docker_init_files.tar.gz
rm -rf root/static/

cp ../service/upcycling-xmas-tree-service usr/bin/
cp -rf ../service/static/ root/static/

tar cvzf docker_init_files.tar.gz etc root usr
