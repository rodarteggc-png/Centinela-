package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"centinela/sampler/internal/config"
	"centinela/sampler/internal/sqlserver"
	mssql "github.com/microsoft/go-mssqldb"
)

func main() {
	inventory := flag.String("inventory", "", "Inventory file to validate")
	names := flag.String("targets", "", "Comma-separated target names to validate")
	flag.Parse()
	if *inventory == "" || *names == "" {
		fmt.Fprintln(os.Stderr, "Both -inventory and -targets are required")
		os.Exit(1)
	}
	source, err := os.Open(*inventory)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot open inventory")
		os.Exit(1)
	}
	targets, err := config.LoadSQLServerTargets(source)
	source.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid inventory")
		os.Exit(1)
	}
	wanted := make(map[string]bool)
	for _, name := range strings.Split(*names, ",") {
		wanted[strings.TrimSpace(name)] = false
	}
	failed := false
	for _, target := range targets {
		if _, ok := wanted[target.Name]; !ok {
			continue
		}
		wanted[target.Name] = true
		if err := check(target); err != nil {
			var sqlError mssql.Error
			if errors.As(err, &sqlError) {
				fmt.Printf("%s FAILED SQL_CODE=%d\n", target.Name, sqlError.Number)
			} else {
				fmt.Printf("%s FAILED connection, timeout, permissions or sampling\n", target.Name)
			}
			failed = true
		}
	}
	for name, found := range wanted {
		if !found {
			fmt.Printf("%s FAILED not found in inventory\n", name)
			failed = true
		}
	}
	if failed {
		os.Exit(1)
	}
}

func check(target config.Target) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	database, err := sql.Open("sqlserver", target.ConnectionString())
	if err != nil {
		return err
	}
	defer database.Close()
	var major, viewState, viewPerformance int
	err = database.QueryRowContext(ctx, `SELECT
		CONVERT(int, SERVERPROPERTY('ProductMajorVersion')),
		COALESCE(HAS_PERMS_BY_NAME(NULL, NULL, 'VIEW SERVER STATE'), 0),
		COALESCE(HAS_PERMS_BY_NAME(NULL, NULL, 'VIEW SERVER PERFORMANCE STATE'), 0)
	`).Scan(&major, &viewState, &viewPerformance)
	if err != nil {
		return err
	}
	if viewState != 1 || (major >= 16 && viewPerformance != 1) {
		return fmt.Errorf("required server DMV permissions missing")
	}
	sampler, err := sqlserver.Open(target)
	if err != nil {
		return err
	}
	defer sampler.Close()
	logical, visible, err := sampler.DetectCapacity(ctx)
	if err != nil {
		return err
	}
	started := time.Now()
	sessions, _, err := sampler.Sample(ctx, target.Name, "preflight")
	if err != nil {
		return err
	}
	fmt.Printf("%s OK SQL_MAJOR=%d VIEW_SERVER_STATE=%d VIEW_SERVER_PERFORMANCE_STATE=%d LOGICAL_CPUS=%d VISIBLE_SCHEDULERS=%d ACTIVE_SESSIONS=%d SAMPLE_MS=%d\n",
		target.Name, major, viewState, viewPerformance, logical, visible, len(sessions), time.Since(started).Milliseconds())
	return nil
}