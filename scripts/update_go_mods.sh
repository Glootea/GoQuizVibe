#!/bin/bash
set -e

echo "=== Update go.mod to use backend module path ==="

cat > go.mod << 'EOF'
module github.com/goquizvibe/backend

go 1.26.0

require (
	github.com/a-h/templ v0.3.1001
	github.com/andybalholm/brotli v1.1.0
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/golang-migrate/migrate/v4 v4.19.1
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.9.2
	github.com/joho/godotenv v1.5.1
	github.com/klauspost/compress v1.18.6
	github.com/minio/minio-go/v7 v7.1.0
	go.uber.org/mock v0.6.0
	golang.org/x/crypto v0.49.0
)

require (
	github.com/goforj/wire v1.2.0
	github.com/goquizvibe/pkg v0.0.0
	github.com/invopop/jsonschema v0.14.0
	github.com/leonelquinteros/gotext v1.7.2
	github.com/prometheus/client_golang v1.23.2
	github.com/redis/go-redis/v9 v9.19.0
)

require (
	cel.dev/expr v0.25.1
	filippo.io/edwards25519 v1.1.1
	github.com/Glootea/gettextgocodegen v0.0.0-20260515195814-005425908ffb
	github.com/a-h/parse v0.0.0-20250122154542-74294addb73e
	github.com/antlr4-go/antlr/v4 v4.13.1
	github.com/bahlo/generic-list-go v0.2.0
	github.com/beorn7/perks v1.0.1
	github.com/buger/jsonparser v1.1.2
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/cespare/xxhash/v2 v2.3.0
	github.com/cli/browser v1.3.0
	github.com/coreos/go-semver v0.3.1
	github.com/cubicdaiya/gonp v1.0.4
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/dustin/go-humanize v1.0.1
	github.com/fatih/color v1.16.0
	github.com/fatih/structtag v1.2.0
	github.com/fsnotify/fsnotify v1.7.0
	github.com/go-ini/ini v1.67.0
	github.com/go-sql-driver/mysql v1.9.3
	github.com/google/cel-go v0.28.0
	github.com/google/subcommands v1.2.0
	github.com/inconshreveable/mousetrap v1.1.0
	github.com/jackc/pgpassfile v1.0.0
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761
	github.com/jackc/puddle/v2 v2.2.2
	github.com/jinzhu/inflection v1.0.0
	github.com/klauspost/cpuid/v2 v2.2.11
	github.com/klauspost/crc32 v1.3.0
	github.com/lib/pq v1.12.3
	github.com/mattn/go-colorable v0.1.13
	github.com/mattn/go-isatty v0.0.20
	github.com/minio/crc64nvme v1.1.1
	github.com/minio/md5-simd v1.1.2
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822
	github.com/natefinch/atomic v1.0.1
	github.com/ncruces/go-sqlite3 v0.32.0
	github.com/ncruces/julianday v1.0.0
	github.com/pb33f/ordered-map/v2 v2.3.1
	github.com/pganalyze/pg_query_go/v6 v6.2.2
	github.com/philhofer/fwd v1.2.0
	github.com/pingcap/errors v0.11.5-0.20250523034308-74f78ae071ee
	github.com/pingcap/failpoint v0.0.0-20240528011301-b51a646c7c86
	github.com/pingcap/log v1.1.0
	github.com/pingcap/tidb/pkg/parser v0.0.0-20260418072757-ce92298d1124
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2
	github.com/prometheus/client_model v0.6.2
	github.com/prometheus/common v0.66.1
	github.com/prometheus/procfs v0.16.1
	github.com/riza-io/grpc-go v0.2.0
	github.com/rogpeppe/go-internal v1.14.1
	github.com/rs/xid v1.6.0
	github.com/spf13/cobra v1.10.2
	github.com/spf13/pflag v1.0.10
	github.com/sqlc-dev/doubleclick v1.0.0
	github.com/sqlc-dev/sqlc v1.31.1
	github.com/tetratelabs/wazero v1.11.0
	github.com/tinylib/msgp v1.6.1
	github.com/wasilibs/go-pgquery v0.0.0-20250409022910-10ac41983c07
	github.com/wasilibs/wazero-helpers v0.0.0-20240620070341-3dff1577cd52
	github.com/zeebo/xxh3 v1.1.0
	go.uber.org/atomic v1.11.0
	go.uber.org/multierr v1.11.0
	go.uber.org/zap v1.27.0
	go.yaml.in/yaml/v2 v2.4.2
	go.yaml.in/yaml/v3 v3.0.4
	go.yaml.in/yaml/v4 v4.0.0-rc.2
	golang.org/x/exp v0.0.0-20250620022241-b7579e27df2b
	golang.org/x/mod v0.34.0
	golang.org/x/net v0.52.0
	golang.org/x/sync v0.20.0
	golang.org/x/sys v0.43.0
	golang.org/x/text v0.36.0
	golang.org/x/tools v0.43.0
	google.golang.org/genproto/googleapis/api v0.0.0-20260120221211-b8f7ae30c516
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260120221211-b8f7ae30c516
	google.golang.org/grpc v1.80.0
	google.golang.org/protobuf v1.36.11
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/goquizvibe/pkg => ../pkg

tool (
	github.com/Glootea/gettextgocodegen
	github.com/a-h/templ/cmd/templ
	github.com/goforj/wire/cmd/wire
	github.com/sqlc-dev/sqlc/cmd/sqlc
)
EOF

echo "=== Update go.work ==="
cat > go.work << 'EOF'
go 1.26.0

use (
	.
	../microservices/typst
)

replace github.com/goquizvibe/pkg => ./pkg
EOF

echo "go.mod and go.work updated"