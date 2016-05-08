# gossip
simple discovery service / secret store written in golang

$ docker build -t gossip .

Dev:
$ docker run --net=host -ti gossip /bin/sh

In production:
$ docker run --net=host --restart=always gossip ./main -pass <PASSWORD> -port <PORT>