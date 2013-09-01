package main

import (
	"github.com/mitchellh/packer/builder/lxc"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	plugin.ServeBuilder(new(lxc.Builder))
}
