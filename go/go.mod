module berty.tech/go

go 1.12

require (
	github.com/brianvoe/gofakeit v3.18.0+incompatible
	github.com/gogo/protobuf v1.3.0
	github.com/golang/protobuf v1.3.2
	github.com/improbable-eng/grpc-web v0.11.0
	github.com/ipfs/go-datastore v0.1.1
	github.com/ipfs/go-ipfs v0.4.22
	github.com/ipfs/go-ipfs-config v0.0.3
	github.com/ipfs/interface-go-ipfs-core v0.0.8
	github.com/jinzhu/gorm v1.9.11
	github.com/libp2p/go-libp2p v0.0.28
	github.com/libp2p/go-libp2p-core v0.0.1
	github.com/libp2p/go-libp2p-crypto v0.1.0
	github.com/libp2p/go-libp2p-peer v0.2.0
	github.com/multiformats/go-multiaddr v0.0.4
	github.com/multiformats/go-multiaddr-net v0.0.1
	github.com/multiformats/go-multihash v0.0.8
	github.com/oklog/run v1.0.0
	github.com/peterbourgon/ff v1.6.0
	github.com/pkg/errors v0.8.1
	github.com/smartystreets/goconvey v0.0.0-20190731233626-505e41936337 // indirect
	go.uber.org/multierr v1.2.0 // indirect
	go.uber.org/zap v1.10.0
	golang.org/x/crypto v0.0.0-20190923035154-9ee001bba392
	golang.org/x/net v0.0.0-20191002035440-2ec189313ef0
	google.golang.org/genproto v0.0.0-20190927181202-20e1ac93f88c
	google.golang.org/grpc v1.24.0
	gopkg.in/gormigrate.v1 v1.6.0
	gopkg.in/yaml.v2 v2.2.3 // indirect
)

replace github.com/libp2p/go-openssl v0.0.2 => github.com/berty/go-openssl v0.0.3-0.20191007152928-66bd988d235e
