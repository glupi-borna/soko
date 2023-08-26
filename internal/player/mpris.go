package player

import (
	"errors"
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

var nilVariantErr = errors.New("Variant value is nil")
var missingKeyErr = errors.New("Missing key")

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
	var names []string

	err := conn.BusObject().Call(dbusListNames, 0).Store(&names)
	if err != nil { return nil, err }

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

func (i *Player) GetName() string { return i.name }
func (i *Player) Raise() error { return i.obj.Call(mprisRaise, 0).Err }
func (i *Player) Quit() error { return i.obj.Call(mprisQuit, 0).Err }
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
	s, err := playerGet[string](i, "PlaybackStatus")
	if err != nil { return PlaybackStatus(""), err }
	return PlaybackStatus(s), nil
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

func (i *Player) GetInfo(name string) (any, error) {
	meta, err := i.GetMetadata()
	if err != nil { return nil, err }
	if meta == nil { return nil, nilVariantErr }
	val, ok := meta[name]
	if !ok { return nil, missingKeyErr }
	return val.Value(), nil
}

func playerInfo[K any](i *Player, name string) (out K, err error) {
	val, err := i.GetInfo(name)
	if err != nil { return out, err }
	out = val.(K)
	return out, nil
}

// Returns the current track length in seconds.
func (i *Player) GetLength() (float64, error) {
	out, err := playerInfo[int64](i, "mpris:length")
	if err != nil { return 0.0, err }
	return usToS(out), nil
}

func (i *Player) GetTrackId() (dbus.ObjectPath, error) { return playerInfo[dbus.ObjectPath](i, "mpris:trackid") }
func (i *Player) GetArtUrl() (string, error) { return playerInfo[string](i, "mpris:artUrl") }
func (i *Player) GetAlbum() (string, error) { return playerInfo[string](i, "xesam:album") }
func (i *Player) GetAlbumArtist() (string, error) { return playerInfo[string](i, "xesam:albumArtist") }
func (i *Player) GetArtists() ([]string, error) { return playerInfo[[]string](i, "xesam:artist") }
func (i *Player) GetLyrics() (string, error) { return playerInfo[string](i, "xesam:asText") }
func (i *Player) GetBPM() (int, error) { return playerInfo[int](i, "xesam:audioBPM") }
func (i *Player) GetTitle() (string, error) { return playerInfo[string](i, "xesam:title") }
func (i *Player) GetUrl() (string, error) { return playerInfo[string](i, "xesam:url") }

type TrackInfo struct {
	TrackId     dbus.ObjectPath
	ArtUrl      string
	Album       string
	AlbumArtist string
	Artists     []string
	Lyrics      string
	BPM         int
	Title       string
	Url         string
}

var track_info_cache = make(map[dbus.ObjectPath]*TrackInfo)

func ignore_err[K any](fn func()(K, error)) K {
	val, _ := fn()
	return val
}

func (i *Player) GetTrackInfo() (*TrackInfo, error) {
	track_id, err := i.GetTrackId()
	if err != nil { return nil, err }

	ti, ok := track_info_cache[track_id]
	if ok { return ti, nil }

	ti = &TrackInfo{
		TrackId: track_id,
		ArtUrl: ignore_err(i.GetArtUrl),
		Album: ignore_err(i.GetAlbum),
		AlbumArtist: ignore_err(i.GetAlbumArtist),
		Artists: ignore_err(i.GetArtists),
		Lyrics: ignore_err(i.GetLyrics),
		BPM: ignore_err(i.GetBPM),
		Title: ignore_err(i.GetTitle),
		Url: ignore_err(i.GetUrl),
	}

	track_info_cache[track_id] = ti
	return ti, nil
}

// Returns the current track position in seconds.
func (i *Player) GetPosition() (float64, error) {
	pos, err := playerGet[int64](i, "Position")
	if err != nil { return 0.0, err }
	return usToS(pos), nil
}

// Sets the position of the current track in seconds.
func (i *Player) SetPosition(position float64) error {
	trackId, err := i.GetTrackId()
	if err != nil { return err }
	i.SetTrackPosition(&trackId, position)
	return nil
}

var WidgetVars = map[string]any{
	"Players": func() ([]*Player, error) {
		conn, err := globals.DbusConn()
		if err != nil { return nil, err }

		player_names, err := Players(conn)
		if err != nil { return nil, err }

		if len(player_names) == 0 {
			return []*Player{}, nil
		}

		out := make([]*Player, 0, len(player_names))
		for _, name := range player_names {
			player := MakePlayer(conn, name)
			out = append(out, player)
		}

		return out, nil
	},
}
