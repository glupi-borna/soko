package globals

import (
	"github.com/godbus/dbus/v5"
)

var Close func()

var dbusConn *dbus.Conn
func DbusConn() (*dbus.Conn, error) {
	if dbusConn == nil {
		var err error
		dbusConn, err = dbus.ConnectSessionBus()
		if err != nil { return nil, err }
	}
	return dbusConn, nil
}
