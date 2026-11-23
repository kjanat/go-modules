module tests

go 1.27.0

toolchain go1.27.4

require (
	github.com/kjanat/go-modules v0.0.0-00010101000000-000000000000
	github.com/kjanat/go-modules/cm v0.0.0-00010101000000-000000000000
	go.bytecodealliance.org/cm v0.3.0
	golang.org/x/tools v0.53.0
)

require (
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/kjanat/wazero v1.12.2 // indirect
	github.com/klauspost/compress v1.20.3 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/regclient/regclient v0.11.7 // indirect
	github.com/sirupsen/logrus v1.10.2 // indirect
	github.com/ulikunitz/xz v0.5.17 // indirect
	github.com/urfave/cli/v3 v3.14.0 // indirect
	golang.org/x/mod v0.43.0 // indirect
	golang.org/x/sync v0.25.0 // indirect
	golang.org/x/sys v0.50.0 // indirect
)

replace github.com/kjanat/go-modules => ../

replace github.com/kjanat/go-modules/cm => ../cm
