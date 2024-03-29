package impl

import (
	"context"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/okta/okta-sdk-golang/okta"
	"github.com/spf13/viper"

	"github.com/netauth/netauth/pkg/plugin/tree"
)

var (
	appLogger hclog.Logger

	cfg *viper.Viper
)

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.netauth")
	viper.AddConfigPath("/etc/netauth/")

	if err := viper.ReadInConfig(); err != nil {
		appLogger.Error("Fatal error reading configuration", "error", err)
		os.Exit(1)
	}

	viper.SetDefault("log.level", "INFO")
	appLogger = hclog.New(&hclog.LoggerOptions{
		Name:  "okta",
		Level: hclog.LevelFromString(viper.GetString("log.level")),
	})
	hclog.SetDefault(appLogger)

	viper.SetDefault("plugin.okta.orgurl", "")
	viper.SetDefault("plugin.okta.tokan", "")
	viper.SetDefault("plugin.okta.domain", "")
	viper.SetDefault("plugin.okta.interval", time.Minute*20)

	viper.Set("client.ServiceName", "okta-groupsync")

	cfg = viper.Sub("plugin.okta")
}

// New will return a plugin fully provisioned and ready to go.
func New() tree.Plugin {
	ctx := context.Background()

	client, err := okta.NewClient(ctx,
		okta.WithOrgUrl(cfg.GetString("orgurl")),
		okta.WithToken(cfg.GetString("token")))

	if err != nil {
		appLogger.Error("Okta Error during initailization", "error", err)
	}

	x := OktaPlugin{
		NullPlugin: tree.NullPlugin{},
		c:          client,
	}

	x.syncGroups()
	go x.groupSyncTimer()

	return x
}
