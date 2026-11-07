package config

import (
	"strings"
	"testing"
	"net/url"
)

func TestLoadSQLServerTargetsSkipsCommentsAndOtherEngines(t *testing.T) {
	inventory := `# generated inventory
engine,instance_name,host_or_ip,port,environment,cluster_name,username,password,database_name
sqlserver,node-1,192.0.2.10,5000,production,cluster,monitor,p@ss$,master
oracle,oracle-1,db.example,1521,development,oracle,monitor,unused,orcl
`

	targets, err := LoadSQLServerTargets(strings.NewReader(inventory))
	if err != nil {
		t.Fatalf("LoadSQLServerTargets returned an error: %v", err)
	}
	if len(targets) != 1 || targets[0].Name != "node-1" {
		t.Fatalf("unexpected targets: %+v", targets)
	}
	connectionURL, err := url.Parse(targets[0].ConnectionString())
	if err != nil {
		t.Fatalf("connection string is invalid: %v", err)
	}
	username := connectionURL.User.Username()
	password, ok := connectionURL.User.Password()
	if !ok || username != targets[0].Username || password != targets[0].Password {
		t.Fatal("connection string did not preserve encoded credentials")
	}
}

func TestLoadSQLServerTargetsRejectsDuplicateNames(t *testing.T) {
	inventory := `engine,instance_name,host_or_ip,port,environment,cluster_name,username,password,database_name
sqlserver,node-1,192.0.2.10,5000,production,cluster,monitor,secret,master
sqlserver,node-1,192.0.2.11,5000,production,cluster,monitor,secret,master
`
	if _, err := LoadSQLServerTargets(strings.NewReader(inventory)); err == nil {
		t.Fatal("expected duplicate instance_name error")
	}
}