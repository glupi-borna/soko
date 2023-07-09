package globals

import (
	"os"
	"strings"
	"github.com/godbus/dbus/v5"
)

var Close func()

func envDef(name string, def string) string {
	res := os.Getenv(name)
	if name == "" { return def }
	return res
}

var dbusConn *dbus.Conn
var Home = os.Getenv("HOME")
var Display = strings.TrimPrefix(os.Getenv("DISPLAY"), ":")
var dbusSessionBusAddress = os.Getenv("DBUS_SESSION_BUS_ADDRESS")
const dbusFileAddressLinePrefix = "DBUS_SESSION_BUS_ADDRESS='"

func dbusPath() string {
	if dbusSessionBusAddress != "" {
		return dbusSessionBusAddress
	}

	if Home == "" {
		panic("Missing $HOME environment variable: can't find home directory!")
	}

	sessionBusFolderPath := Home + "/.dbus" + "/session-bus"
	sessionBusFolder, err := os.Stat(sessionBusFolderPath)
	if err != nil { panic(err.Error()) }

	if !sessionBusFolder.IsDir() {
		panic("$HOME/.dbus/session-bus is not a directory!")
	}

	dbusFilePath := ""
	ents, err := os.ReadDir(sessionBusFolderPath)
	if err != nil { panic(err.Error()) }
	for _, ent := range ents {
		if strings.HasSuffix(ent.Name(), Display) {
			dbusFilePath = sessionBusFolderPath + "/" + ent.Name()
		}
	}

	if dbusFilePath == "" {
		panic("No dbus config file exists in $HOME/.dbus/session-bus for display " + Display)
	}

	bytes, err := os.ReadFile(dbusFilePath)
	if err != nil { panic(err.Error()) }

	lines := strings.Split(string(bytes), "\n")
	addressLine := ""
	for _, line := range lines {
		if strings.HasPrefix(line, dbusFileAddressLinePrefix) {
			addressLine = line
			break
		}
	}

	if addressLine == "" {
		panic("dbus config has no address variable: " + dbusFilePath)
	}

	address := addressLine[len(dbusFileAddressLinePrefix):len(addressLine)-1]
	return address
}

func DbusConn() (*dbus.Conn, error) {
	os.Setenv("DBUS_SESSION_BUS_ADDRESS", dbusPath())
	if dbusConn == nil {
		var err error
		dbusConn, err = dbus.ConnectSessionBus()
		if err != nil { return nil, err }
	}
	return dbusConn, nil
}
