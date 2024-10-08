module github.com/dailycrypto-me/daily-evm

go 1.18

require (
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/davecgh/go-spew v1.1.1
	github.com/emicklei/dot v1.2.0
	github.com/holiman/uint256 v1.2.1
	github.com/linxGnu/grocksdb v1.6.48
	github.com/otiai10/copy v1.9.0
	github.com/schollz/progressbar/v3 v3.3.3
	github.com/stretchr/testify v1.8.1
	golang.org/x/crypto v0.3.0
	golang.org/x/sys v0.2.0
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f
)

require (
	github.com/kr/text v0.1.0 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/linxGnu/grocksdb v1.6.48 => github.com/Taraxa-project/grocksdb v1.6.48-taraxa
