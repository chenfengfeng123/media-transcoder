module github.com/harisbeha/media-transcoder

go 1.13

require (
	cloud.google.com/go v0.44.3
	github.com/aws/aws-sdk-go v1.23.8
	github.com/gocraft/work v0.5.1
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/jmoiron/sqlx v1.2.0
	github.com/labstack/echo/v4 v4.1.10
	github.com/lib/pq v1.1.1
	github.com/robfig/cron v1.2.0 // indirect
	github.com/rs/xid v1.2.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	google.golang.org/api v0.8.0
	k8s.io/api v0.0.0-20190905160310-fb749d2f1064
	k8s.io/apimachinery v0.0.0-20190831074630-461753078381
	k8s.io/client-go v0.0.0-20190906195228-67a413f31aea
	k8s.io/utils v0.0.0-20190801114015-581e00157fb1
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190905160310-fb749d2f1064
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190831074630-461753078381
)
