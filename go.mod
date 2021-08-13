module github.com/meano/drone-runner-native

go 1.16

replace github.com/drone/runner-go => github.com/Meano/drone-runner-go v1.8.1-0.20210813052805-4d72fc7a7456

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20210208195552-ff826a37aa15 // indirect
	github.com/buildkite/yaml v2.1.0+incompatible
	github.com/dchest/uniuri v0.0.0-20200228104902-7aecb25e1fe5
	github.com/drone/drone-go v1.6.0
	github.com/drone/drone-yaml v1.2.3
	github.com/drone/runner-go v1.8.0
	github.com/drone/signal v1.0.0
	github.com/ghodss/yaml v1.0.0
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/joho/godotenv v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0 // indirect
	golang.org/x/net v0.0.0-20210805182204-aaa1db679c0d // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210809222454-d867a43fc93e // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
)
