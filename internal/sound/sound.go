package sound

import (
	"time"
	"os"
	"fmt"
	"strings"
	"errors"
	"math"
	pa "mrogalski.eu/go/pulseaudio"
)

func Die(err error) {
	if err != nil { panic(err.Error()) }
}

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
		fmt.Println("Default socket at", def, "does not exist, falling back to /tmp/")
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

var pulseClient *pa.Client = nil

func pulseInit() error {
	if pulseClient != nil { return nil }
	socket, err := getPulseSocket()
	if err != nil { return err }
	pulseClient, err = pa.NewClient(socket)
	if err != nil { return err }
	return nil
}

// Returns the current PC volume in 0-1 range (values above
// 1 mean that volume is boosted).
func Volume() (float32, error) {
	err := pulseInit()
	if err != nil { return float32(math.NaN()), err }
	return pulseClient.Volume()
}
