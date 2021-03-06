package digitalocean

import (
	"cgl.tideland.biz/identifier"
	"encoding/hex"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"time"
)

type stepCreateDroplet struct {
	dropletId uint
}

func (s *stepCreateDroplet) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*DigitalOceanClient)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(config)
	sshKeyId := state.Get("ssh_key_id").(uint)

	ui.Say("Creating droplet...")

	// Some random droplet name as it's temporary
	name := fmt.Sprintf("packer-%s", hex.EncodeToString(identifier.NewUUID().Raw()))

	// Create the droplet based on configuration
	dropletId, err := client.CreateDroplet(name, c.SizeID, c.ImageID, c.RegionID, sshKeyId)
	if err != nil {
		err := fmt.Errorf("Error creating droplet: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// We use this in cleanup
	s.dropletId = dropletId

	// Store the droplet id for later
	state.Put("droplet_id", dropletId)

	return multistep.ActionContinue
}

func (s *stepCreateDroplet) Cleanup(state multistep.StateBag) {
	// If the dropletid isn't there, we probably never created it
	if s.dropletId == 0 {
		return
	}

	client := state.Get("client").(*DigitalOceanClient)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(config)

	// Destroy the droplet we just created
	ui.Say("Destroying droplet...")

	// Sleep arbitrarily before sending destroy request
	// Otherwise we get "pending event" errors, even though there isn't
	// one.
	log.Printf("Sleeping for %v, event_delay", c.RawEventDelay)
	time.Sleep(c.eventDelay)

	var err error
	for i := 0; i < 5; i++ {
		err = client.DestroyDroplet(s.dropletId)
		if err == nil {
			break
		}

		time.Sleep(2 * time.Second)
	}

	if err != nil {
		curlstr := fmt.Sprintf("curl '%v/droplets/%v/destroy?client_id=%v&api_key=%v'",
			DIGITALOCEAN_API_URL, s.dropletId, c.ClientID, c.APIKey)

		ui.Error(fmt.Sprintf(
			"Error destroying droplet. Please destroy it manually: %v", curlstr))
	}
}
