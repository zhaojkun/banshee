#!/bin/bash -e
cd /banshee
if [ -f data/config.yaml ]
then
    ./banshee -c data/config.yaml
else
    ./banshee -c config.yaml
fi    
