module github.com/bogdanrat/web-server/service/core

go 1.16

require (
	cloud.google.com/go v0.89.0 // indirect
	github.com/aws/aws-sdk-go v1.40.13
	github.com/bogdanrat/web-server/contracts v0.0.0-20210803173554-07da162fcd1b
	github.com/bogdanrat/web-server/service/monitor v0.0.0-20210803173554-07da162fcd1b
	github.com/bogdanrat/web-server/service/queue v0.0.0-20210803173554-07da162fcd1b
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-gonic/gin v1.7.3
	github.com/go-playground/validator/v10 v10.8.0 // indirect
	github.com/go-redis/redis/v7 v7.4.1
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/lib/pq v1.10.2
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/common v0.30.0 // indirect
	github.com/prometheus/procfs v0.7.1 // indirect
	github.com/spf13/cast v1.4.0 // indirect
	github.com/spf13/viper v1.8.1
	github.com/streadway/amqp v1.0.0
	github.com/ugorji/go v1.2.6 // indirect
	golang.org/x/net v0.0.0-20210726213435-c6fcb2dbf985 // indirect
	golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914
	google.golang.org/api v0.52.0
	google.golang.org/genproto v0.0.0-20210803142424-70bd63adacf2
	google.golang.org/grpc v1.39.0
)
