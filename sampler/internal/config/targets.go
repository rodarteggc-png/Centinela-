package config

import (
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
)

type Target struct {
	Name        string
	Host        string
	Port        uint16
	Environment string
	Cluster     string
	Username    string
	Password    string
	Database    string
}

func LoadSQLServerTargets(source io.Reader) ([]Target, error) {
	contents, err := io.ReadAll(source)
	if err != nil {
		return nil, fmt.Errorf("read inventory: %w", err)
	}
	lines := strings.Split(string(contents), "\n")
	filteredLines := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		filteredLines = append(filteredLines, line)
	}
	reader := csv.NewReader(strings.NewReader(strings.Join(filteredLines, "\n")))
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read inventory header: %w", err)
	}
	columns := make(map[string]int, len(header))
	for index, name := range header {
		columns[strings.TrimSpace(name)] = index
	}
	for _, required := range []string{"engine", "instance_name", "host_or_ip", "port", "environment", "cluster_name", "username", "password", "database_name"} {
		if _, ok := columns[required]; !ok {
			return nil, fmt.Errorf("inventory is missing column %q", required)
		}
	}

	var targets []Target
	seen := make(map[string]struct{})
	for rowNumber := 2; ; rowNumber++ {
		row, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("read inventory row %d: %w", rowNumber, readErr)
		}
		if len(row) == 0 || strings.TrimSpace(strings.Join(row, "")) == "" {
			continue
		}
		value := func(column string) string { return strings.TrimSpace(row[columns[column]]) }
		if strings.ToLower(value("engine")) != "sqlserver" {
			continue
		}
		name := value("instance_name")
		if name == "" {
			return nil, fmt.Errorf("inventory row %d has an empty instance_name", rowNumber)
		}
		if _, exists := seen[name]; exists {
			return nil, fmt.Errorf("duplicate SQL Server instance_name %q", name)
		}
		seen[name] = struct{}{}

		port, parseErr := strconv.ParseUint(value("port"), 10, 16)
		if parseErr != nil || port == 0 {
			return nil, fmt.Errorf("inventory row %d has an invalid port", rowNumber)
		}
		target := Target{
			Name:        name,
			Host:        value("host_or_ip"),
			Port:        uint16(port),
			Environment: value("environment"),
			Cluster:     value("cluster_name"),
			Username:    value("username"),
			Password:    value("password"),
			Database:    value("database_name"),
		}
		if target.Host == "" || target.Username == "" || target.Password == "" || target.Database == "" {
			return nil, fmt.Errorf("inventory row %d is missing connection fields", rowNumber)
		}
		targets = append(targets, target)
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("inventory has no SQL Server targets")
	}
	return targets, nil
}

func (target Target) ConnectionString() string {
	connectionURL := url.URL{
		Scheme: "sqlserver",
		User:   url.UserPassword(target.Username, target.Password),
		Host:   net.JoinHostPort(target.Host, strconv.Itoa(int(target.Port))),
	}
	query := connectionURL.Query()
	query.Set("database", target.Database)
	query.Set("encrypt", "disable")
	connectionURL.RawQuery = query.Encode()
	return connectionURL.String()
}
