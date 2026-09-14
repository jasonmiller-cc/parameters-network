module github.com/jasonmiller-cc/parameters-network

go 1.27.1

require (
	github.com/jasonmiller-cc/parameters-core v0.0.0
	github.com/vishvananda/netlink v1.3.0
	golang.org/x/sys v0.47.0
)

require (
	github.com/vishvananda/netns v0.0.4 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/jasonmiller-cc/parameters-core => ../parameters-core
