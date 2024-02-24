#!/bin/bash

cd tofiks
git pull
make build
systemctl stop tofiks
cp tofiks ~/lichess-bot/engines/tofiks
systemctl start tofiks
exit