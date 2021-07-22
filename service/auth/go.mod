module github.com/bogdanrat/web-server/service/auth

go 1.16

require (
	github.com/bogdanrat/web-server v0.0.0-20210524125729-a39d25dcba12
	github.com/bogdanrat/web-server/contracts v0.0.0-20210525184813-14ea474ff934
	github.com/bogdanrat/web-server/service/monitor v0.0.0-20210722123852-4e5407f87ed3
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dgryski/dgoogauth v0.0.0-20190221195224-5a805980a5f3
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-gonic/gin v1.7.2
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1 // indirect
	github.com/twinj/uuid v1.0.0
	go.opencensus.io v0.22.0
	google.golang.org/genproto v0.0.0-20210524171403-669157292da3
	google.golang.org/grpc v1.38.0
	rsc.io/qr v0.2.0
)
