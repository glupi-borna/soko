package sound

import (
	"time"
	"os"
	"fmt"
	"strings"
	"errors"
	pa "mrogalski.eu/go/pulseaudio"
	. "github.com/glupi-borna/wiggo/internal/utils"
)

// @TODO: Consider forking https://github.com/mafik/pulseaudio/ to add more features here.
// @NOTE: https://github.com/mafik/pulseaudio/ has a memory leak in the `bread` function

var pulseClient *pa.Client = nil

func getDefaultPulseSocket() (string, error) {
	path := fmt.Sprintf("/run/user/%d/pulse/native", os.Getuid())
	_, err := os.Stat(path)
	if err != nil { return path, err }
	return path, nil
}

func getPulseSocket() (string, error) {
	def, err := getDefaultPulseSocket()
	if err == nil { return def, nil }

	if err != nil {
		fmt.Println("Pulse: Default socket at", def, "does not exist, falling back to /tmp/")
	}

	files, err := os.ReadDir("/tmp")
	Die(err)

	var latest string = ""
	var latest_time = time.Time{}
	for _, ent := range files {
		if !strings.HasPrefix(ent.Name(), "pulse-") { continue }
		if !ent.IsDir() { continue }
		path := "/tmp/" + ent.Name() + "/native"
		stat, err := os.Stat(path)
		if err != nil { continue }
		mod_time := stat.ModTime()
		if mod_time.After(latest_time) {
			latest = path
			latest_time = mod_time
		}
	}

	if latest == "" {
		return "", errors.New("Failed to find pulseaudio socket in /tmp!")
	}

	return latest, nil
}

func pulseInit() error {
	if pulseClient != nil { return nil }
	socket, err := getPulseSocket()
	if err != nil { return err }
	pulseClient, err = pa.NewClient(socket)
	if err != nil { return err }
	return nil
}

// Returns the volume of the default sink in 0-1 range.
// Values above 1 mean that the volume is above 100%.
func Volume() (float32, error) {
	err := pulseInit()
	if err != nil { return NaN32, err }
	return pulseClient.Volume()
}

// Sets the current volume of the default sink.
// 0-1 range, values above 1 will boost the volume
// above 100%.
func SetVolume(val float32) error {
	err := pulseInit()
	if err != nil { return err }
	return pulseClient.SetVolume(val)
}

// Checks if the default sink is muted.
func IsMuted() (bool, error) {
	err := pulseInit()
	if err != nil { return false, err }
	return pulseClient.Mute()
}

// Sets the muted state of the default sink.
func SetMute(val bool) error {
	err := pulseInit()
	if err != nil { return err }
	return pulseClient.SetMute(val)
}

// Toggles the muted state of the default sink.
func ToggleMute() (bool, error) {
	err := pulseInit()
	if err != nil { return false, err }
	return pulseClient.ToggleMute()
}

var WidgetFns = map[string]any{
	"Volume": Volume,
	"SetVolume": SetVolume,
	"IsMuted": IsMuted,
	"SetMute": SetMute,
	"ToggleMute": ToggleMute,
}
