# gossip
simple discovery service / secret store written in golang

$ docker build -t gossip .

Dev:
$ docker run --net=host -ti -v /c/Users/kamil/dev/gossip:/app/gossip gossip /bin/sh

In production:
$ docker run --net=host --restart=always -d kamilmac/gossip ./main -pass PASSWORD -port PORT