package main

import (
	"github.com/appelgriebsch/terraform-provider-ad/ad"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: ad.Provider})
}
