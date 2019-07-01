module qilin-api

require (
	github.com/ProtocolONE/authone-jwt-verifier-golang v0.0.0-20190415120635-9cfb6c93ff5e
	github.com/ProtocolONE/qilin-common v0.0.0-20190701113525-f0d22a814da0
	github.com/ProtocolONE/rabbitmq v0.0.0-20190129162844-9f24367e139c
	github.com/ProtocolONE/rbac v0.0.0-20190520124240-22f5d2b74988
	github.com/casbin/redis-adapter v0.0.0-20190105032110-b36d844dade5
	github.com/centrifugal/gocent v2.0.2+incompatible
	github.com/denisenkom/go-mssqldb v0.0.0-20181014144952-4e0d7dc8888f // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/erikstmartin/go-testdb v0.0.0-20160219214506-8d10e4a1bae5 // indirect
	github.com/go-sql-driver/mysql v1.4.1 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/gofrs/uuid v3.2.0+incompatible // indirect
	github.com/gogo/protobuf v1.2.1
	github.com/golang/protobuf v1.3.1
	github.com/golang/snappy v0.0.1 // indirect
	github.com/gomodule/redigo v2.0.0+incompatible // indirect
	github.com/jinzhu/gorm v1.9.2
	github.com/jinzhu/inflection v0.0.0-20180308033659-04140366298a // indirect
	github.com/jinzhu/now v0.0.0-20181116074157-8ec929ed50c3 // indirect
	github.com/kelseyhightower/envconfig v1.3.0
	github.com/labstack/echo/v4 v4.0.0
	github.com/labstack/gommon v0.2.8
	github.com/lib/pq v1.0.0
	github.com/lunny/html2md v0.0.0-20181018071239-7d234de44546
	github.com/mattn/go-sqlite3 v1.10.0 // indirect
	github.com/micro/go-micro v1.7.0
	github.com/microcosm-cc/bluemonday v1.0.2
	github.com/mitchellh/mapstructure v1.1.2
	github.com/nats-io/nats-server/v2 v2.0.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/satori/go.uuid v1.2.0
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.3.0
	github.com/tidwall/pretty v1.0.0 // indirect
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c // indirect
	github.com/xdg/stringprep v1.0.0 // indirect
	go.mongodb.org/mongo-driver v1.0.3
	go.uber.org/atomic v1.3.2 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.9.1
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/go-playground/validator.v9 v9.29.0
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	gopkg.in/russross/blackfriday.v2 v2.0.1
)

replace gopkg.in/russross/blackfriday.v2 v2.0.1 => github.com/russross/blackfriday/v2 v2.0.1
