module mc-world-fixer

go 1.23.3

require github.com/OmineDev/neomega-core v0.0.4

require (
	atomicgo.dev/cursor v0.2.0 // indirect
	atomicgo.dev/keyboard v0.2.9 // indirect
	atomicgo.dev/schedule v0.1.0 // indirect
	github.com/containerd/console v1.0.3 // indirect
	github.com/gookit/color v1.5.4 // indirect
	github.com/lithammer/fuzzysearch v1.1.8 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	golang.org/x/sys v0.27.0 // indirect
	golang.org/x/term v0.26.0 // indirect
	golang.org/x/text v0.24.0 // indirect
)

require (
	github.com/df-mc/goleveldb v1.1.9
	github.com/go-gl/mathgl v1.2.0 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pterm/pterm v0.12.80
	github.com/ugorji/go/codec v1.2.12 // indirect
	golang.org/x/image v0.26.0 // indirect
	neomega_blocks v0.0.0 // indirect
)

replace (
	github.com/OmineDev/neomega-core => ../neomega-core
	neomega_blocks => ../neomega-blocks
)
