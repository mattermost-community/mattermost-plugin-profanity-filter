module github.com/mattermost/mattermost-plugin-profanity-filter/server

go 1.12

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b // indirect
	github.com/golang/protobuf v1.1.0 // indirect
	github.com/gorilla/websocket v0.0.0-20180605202552-5ed622c449da // indirect
	github.com/hashicorp/go-hclog v0.0.0-20180402200405-69ff559dc25f // indirect
	github.com/hashicorp/go-plugin v0.0.0-20180331002553-e8d22c780116 // indirect
	github.com/hashicorp/yamux v0.0.0-20180604194846-3520598351bb // indirect
	github.com/mattermost/mattermost-server v5.4.0+incompatible
	github.com/mattermost/viper v1.0.4 // indirect
	github.com/mitchellh/go-testing-interface v0.0.0-20171004221916-a61a99592b77 // indirect
	github.com/nicksnyder/go-i18n v1.10.0 // indirect
	github.com/oklog/run v1.0.0 // indirect
	github.com/pborman/uuid v0.0.0-20170612153648-e790cca94e6c // indirect
	github.com/pkg/errors v0.8.0
	go.uber.org/atomic v1.3.2 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.8.0 // indirect
	golang.org/x/net v0.0.0-20180706051357-32a936f46389 // indirect
	golang.org/x/sync v0.0.0-20190423024810-112230192c58 // indirect
	google.golang.org/genproto v0.0.0-20180627194029-ff3583edef7d // indirect
	google.golang.org/grpc v1.13.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0-20170531160350-a96e63847dc3 // indirect
)

// Workaround for https://github.com/golang/go/issues/30831 and fallout.
replace github.com/golang/lint => github.com/golang/lint v0.0.0-20190227174305-8f45f776aaf1
