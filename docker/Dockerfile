FROM golang:latest
     
RUN apt-get update && \
    apt-get install xz-utils && \
    curl -SLO "https://nodejs.org/dist/v5.12.0/node-v5.12.0-linux-x64.tar.xz" && \
    tar -xJf "node-v5.12.0-linux-x64.tar.xz" && \
    mv node-v5.12.0-linux-x64 /tmp/nodejs && \
    export PATH=$PATH:/tmp/nodejs/bin && \
    npm install -g gulp && \
    git clone https://github.com/eleme/banshee.git $GOPATH/src/github.com/eleme/banshee && \
    cd $GOPATH/src/github.com/eleme/banshee && \
    make static && \
    make build && \
    mkdir -p /banshee/static && \
    cp banshee /banshee &&\
    cp -r  static/dist /banshee/static && \
    cd alerter/exampleCommand && \
    go build &&\
    cp exampleCommand /banshee


VOLUME ["/banshee/data"]

EXPOSE 2015
EXPOSE 2016

COPY ./run.sh /run.sh
COPY config.yaml /banshee/config.yaml

ENTRYPOINT ["/run.sh"]
