package tests

import (
	"testing"
	"mysql_replication_listener"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
//	"fmt"
	"os"
)

func TestStatementReplication(t *testing.T) {
	newConnection := mysql_replication_listener.NewConnection()
	serverId := uint32(2)
	err := newConnection.ConnectAndAuth(HOST, PORT, REPLICATION_USERNAME, REPLICATION_PASSWORD)

	if err != nil {
		t.Fatal("Client not connected and not autentificate to master server with error:", err.Error())
	}
	pos, filename, err := newConnection.GetMasterStatus()

	if err != nil {
		t.Fatal("Master status fail: ", err.Error())
	}

	el, err := newConnection.StartBinlogDump(pos, filename, serverId)

	if err != nil {
		t.Fatal("Cant start bin log: ", err.Error())
	}
	events := el.GetEventChan()

	go func() {
		con, err := sql.Open("mysql", ROOT_USERNAME+":"+ROOT_PASSWORD+"@tcp(localhost:3307)/"+DATABASE)
		defer con.Close()
		if err != nil {
			t.Fatal(err)
		}

		con.Exec("INSERT INTO new_table(text_field, num_field) values(?,?)", "Hello!", 10)

		if (<- events).(*mysql_replication_listener.QueryEvent).GetQuery() != "BEGIN" {
			t.Fatal("Got incorrect query")
		}

		if (<- events).(*mysql_replication_listener.IntVarEvent).GetValue() != 1 {
			t.Fatal("Got incorrect IntEvent")
		}

		expectedQuery := "INSERT INTO new_table(text_field, num_field) values('Hello!',10)"

		if expectedQuery != (<-events).(*mysql_replication_listener.QueryEvent).GetQuery() {
			t.Fatal("Got incorrect query")
		}

		os.Exit(0)

	}()

	err = el.Start()

	if err != nil {
		t.Fatal("Start error", err)
	}
}