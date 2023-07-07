package player

import (
	"fmt"
	"strings"
	"github.com/godbus/dbus/v5"
	"github.com/glupi-borna/soko/internal/globals"
)

const dbusPath = "/org/mpris/MediaPlayer2"

const dbusRoot = "org.freedesktop.DBus"
const dbusListNames = dbusRoot + ".ListNames"
const dbusProps = dbusRoot + ".Properties"
const dbusPropsChanged = dbusProps + ".PropertiesChanged"
const dbusGetProp = dbusProps + ".Get"
const dbusSetProp = dbusProps + ".Set"

const mpris = "org.mpris.MediaPlayer2"
const mprisPlayer = mpris + ".Player"
const mprisTrackList = mpris + ".TrackList"
const mprisPlaylists = mpris + ".Playlists"
const mprisRaise = mpris + ".Raise"
const mprisQuit = mpris + ".Quit"

const playerNext = mprisPlayer + ".Next"
const playerPrevious = mprisPlayer + ".Previous"
const playerPlay = mprisPlayer + ".Play"
const playerPause = mprisPlayer + ".Pause"
const playerPlayPause = mprisPlayer + ".PlayPause"
const playerStop = mprisPlayer + ".Stop"
const playerSeek = mprisPlayer + ".Seek"
const playerSetPosition = mprisPlayer + ".SetPosition"
const playerOpenURI = mprisPlayer + ".OpenURI"
const playerPlaybackStatus = mprisPlayer + ".PlaybackStatus"
const playerLoopStatus = mprisPlayer + ".LoopStatus"

var nilVariantErr = fmt.Errorf("Variant value is nil")

type PlaybackStatus string
const (
	PlaybackPlaying PlaybackStatus = "Playing"
	PlaybackPaused  PlaybackStatus = "Paused"
	PlaybackStopped PlaybackStatus = "Stopped"
)

type LoopStatus string
const (
	LoopNone     LoopStatus = "None"
	LoopTrack    LoopStatus = "Track"
	LoopPlaylist LoopStatus = "Playlist"
)

func sToUs(seconds float64) int64 {
	return int64(seconds * 1000000)
}

func usToS(microseconds int64) float64 {
	return float64(microseconds) / 1000000.0
}

func Players(conn *dbus.Conn) ([]string, error) {
	var mprisNames []string
	names := conn.Names()
	for _, name := range names {
		if strings.HasPrefix(name, mpris) {
			mprisNames = append(mprisNames, name)
		}
	}
	return mprisNames, nil
}

type Player struct {
	conn *dbus.Conn
	obj  *dbus.Object
	name string
}

func MakePlayer(conn *dbus.Conn, name string) *Player {
	obj := conn.Object(name, dbusPath).(*dbus.Object)
	return &Player{conn, obj, name}
}

func (i *Player) Set(prop string, value interface{}) error {
	return i.obj.SetProperty(mprisPlayer + "." + prop, value)
}

func (i *Player) Get(prop string) (dbus.Variant, error) {
	return i.obj.GetProperty(mprisPlayer + "." + prop)
}

func (i *Player) Call(method string, flags dbus.Flags, args... any) error {
	return i.obj.Call(mprisPlayer + "." + method, flags, args...).Err
}

func playerGet[K any](i *Player, prop string) (out K, err error) {
	variant, err := i.Get(prop)

	if err != nil { return out, err }

	val := variant.Value()
	if val == nil { return out, nilVariantErr }

	out = val.(K)
	return out, nil
}

func (i *Player) GetName() string {
	return i.name
}

func (i *Player) Raise() error {
	return i.obj.Call(mprisRaise, 0).Err
}

func (i *Player) Quit() error {
	return i.obj.Call(mprisQuit, 0).Err
}

func (i *Player) Next() error { return i.Call("Next", 0) }
func (i *Player) Previous() error { return i.Call("Previous", 0) }
func (i *Player) Pause() error { return i.Call("Pause", 0) }
func (i *Player) Play() error { return i.Call("Play", 0) }
func (i *Player) PlayPause() error { return i.Call("PlayPause", 0) }
func (i *Player) Stop() error { return i.Call("Stop", 0) }
func (i *Player) Seek(seconds float64) error { return i.Call("Seek", 0, sToUs(seconds)) }
func (i *Player) OpenURI(uri string) error { return i.Call("OpenUri", 0, uri) }
func (i *Player) SetTrackPosition(trackId *dbus.ObjectPath, seconds float64) error {
	return i.Call("SetPosition", 0, trackId, sToUs(seconds))
}

func (i *Player) GetIdentity() (string, error) {
	return playerGet[string](i, "Identity")
}

func (i *Player) GetPlaybackStatus() (PlaybackStatus, error) {
	return playerGet[PlaybackStatus](i, "PlaybackStatus")
}

func (i *Player) GetLoopStatus() (LoopStatus, error) {
	return playerGet[LoopStatus](i, "LoopStatus")
}

func (i *Player) SetLoopStatus(loopStatus LoopStatus) error {
	return i.Set("LoopStatus", loopStatus)
}

func (i *Player) GetRate() (float64, error) {
	return playerGet[float64](i, "Rate")
}

func (i *Player) GetShuffle() (bool, error) {
	return playerGet[bool](i, "Shuffle")
}

func (i *Player) SetShuffle(value bool) error {
	return i.Set("Shuffle", value)
}

func (i *Player) GetMetadata() (map[string]dbus.Variant, error) {
	return playerGet[map[string]dbus.Variant](i, "Metadata")
}

func (i *Player) GetVolume() (float64, error) {
	return playerGet[float64](i, "Volume")
}

func (i *Player) SetVolume(volume float64) error {
	return i.Set("Volume", volume)
}

// Returns the current track length in seconds.
func (i *Player) GetLength() (float64, error) {
	metadata, err := i.GetMetadata()
	if err != nil { return 0.0, err }
	if metadata == nil || metadata["mpris:length"].Value() == nil {
		return 0.0, nilVariantErr
	}
	return usToS(metadata["mpris:length"].Value().(int64)), nil
}

// Returns the current track position in seconds.
func (i *Player) GetPosition() (float64, error) {
	pos, err := playerGet[int64](i, "Position")
	if err != nil { return 0.0, err }
	return usToS(pos), nil
}

// Sets the position of the current track in seconds.
func (i *Player) SetPosition(position float64) error {
	metadata, err := i.GetMetadata()
	if err != nil { return err }
	if metadata == nil || metadata["mpris:trackid"].Value() == nil { return nilVariantErr }
	trackId := metadata["mpris:trackid"].Value().(dbus.ObjectPath)
	i.SetTrackPosition(&trackId, position)
	return nil
}

var WidgetFns = map[string]any{
	"Players": func() ([]*Player, error) {
		conn, err := globals.DbusConn()
		if err != nil { return nil, err }

		player_names, err := Players(conn)
		if err != nil { return nil, err }

		if len(player_names) == 0 {
			return []*Player{}, nil
		}

		out := make([]*Player, len(player_names))
		for _, name := range player_names {
			player := MakePlayer(conn, name)
			out = append(out, player)
		}

		return out, nil
	},
}
