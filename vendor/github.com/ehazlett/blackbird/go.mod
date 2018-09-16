module github.com/ehazlett/blackbird

require (
	github.com/codegangsta/cli v0.0.0-20180821064027-934abfb2f102
	github.com/gogo/protobuf v0.0.0-20180914054005-e14cafb6a2c2
	github.com/mholt/caddy v0.0.0-20180907212407-13a54dbdda25
	github.com/pkg/errors v0.8.0
	github.com/russross/blackfriday v0.0.0-20180829180401-f1f45ab762c2 // indirect
	github.com/sirupsen/logrus v1.0.6
	golang.org/x/net v0.0.0-20180911220305-26e67e76b6c3
	google.golang.org/genproto v0.0.0-20180914223249-4b56f30a1fd9 // indirect
	google.golang.org/grpc v0.0.0-20180914155713-f2aaa9bf7445
)

replace github.com/mholt/caddy => github.com/ehazlett/caddy v0.11.1-0.20180916025321-389f646b27c3
