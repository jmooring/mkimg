module github.com/jmooring/mkimg

go 1.27.2

require (
	github.com/HugoSmits86/nativewebp v1.3.0
	github.com/rogpeppe/go-internal v1.16.0
	golang.org/x/image v0.47.0
)

require (
	github.com/BurntSushi/toml v1.4.1-0.20240526193622-a339e1f7089c // indirect
	golang.org/x/exp/typeparams v0.0.0-20231108232855-2478ac86f678 // indirect
	golang.org/x/mod v0.41.0 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/telemetry v0.0.0-20260924152758-ed294f943157 // indirect
	golang.org/x/tools v0.51.0 // indirect
	golang.org/x/vuln v1.8.0 // indirect
	honnef.co/go/tools v0.8.1 // indirect
)

tool (
	golang.org/x/vuln/cmd/govulncheck
	honnef.co/go/tools/cmd/staticcheck
)
