package main

import (
	"github.com/hepapi/steampipe-plugin-azuredevops-self-hosted/azuredevops"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		PluginFunc: azuredevops.Plugin,
	})
}
