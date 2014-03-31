package vagrant

import (
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"path/filepath"
)

type LxcProvider struct{}

func (p *LxcProvider) KeepInputArtifact() bool {
	return false
}

func (p *LxcProvider) Process(ui packer.Ui, artifact packer.Artifact, dir string) (vagrantfile string, metadata map[string]interface{}, err error) {
	// Create the metadata
	metadata = map[string]interface{}{"provider": "lxc", "version": "1.0.0", "built-on": ""}

	for _, path := range artifact.Files() {
		ui.Message(fmt.Sprintf("Copying: %s", path))

		dstPath := filepath.Join(dir, filepath.Base(path))
		if err = CopyContents(dstPath, path); err != nil {
			return
		}
	}

	vagrantfile = lxcVagrantfile

	return
}



type LxcBoxConfig struct {
	common.PackerConfig `mapstructure:",squash"`

	OutputPath          string `mapstructure:"output"`
	VagrantfileTemplate string `mapstructure:"vagrantfile_template"`

	tpl *packer.ConfigTemplate
}

type LxcBoxPostProcessor struct {
	config LxcBoxConfig
}

var lxcVagrantfile = `
Vagrant.configure("2") do |config|
end
`
