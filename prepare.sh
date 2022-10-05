#!/bin/bash
rm data.bin > /dev/null
rm data.bin.gz > /dev/null
wget https://github.com/noctarius/branded-workshop/raw/main/data.bin.gz
gzip -d data.bin
