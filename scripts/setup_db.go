// +build ignore

// Setup script to create Spanner database for local development.
// Run with: go run scripts/setup_db.go
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	database "cloud.google.com/go/spanner/admin/database/apiv1"
	instance "cloud.google.com/go/spanner/admin/instance/apiv1"
	"cloud.google.com/go/spanner/admin/instance/apiv1/instancepb"
	databasepb "cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
)

const (
	projectID  = "test-project"
	instanceID = "test-instance"
	databaseID = "test-database"
)

func main() {
	ctx := context.Background()

	// Ensure emulator host is set
	if os.Getenv("SPANNER_EMULATOR_HOST") == "" {
		os.Setenv("SPANNER_EMULATOR_HOST", "localhost:9010")
	}

	instanceAdmin, err := instance.NewInstanceAdminClient(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create instance admin client: %v\n", err)
		os.Exit(1)
	}
	defer instanceAdmin.Close()

	dbAdmin, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create database admin client: %v\n", err)
		os.Exit(1)
	}
	defer dbAdmin.Close()

	instanceName := fmt.Sprintf("projects/%s/instances/%s", projectID, instanceID)
	databaseName := fmt.Sprintf("%s/databases/%s", instanceName, databaseID)

	// Try to create instance (ignore if exists)
	fmt.Println("Creating Spanner instance (if not exists)...")
	op, err := instanceAdmin.CreateInstance(ctx, &instancepb.CreateInstanceRequest{
		Parent:     fmt.Sprintf("projects/%s", projectID),
		InstanceId: instanceID,
		Instance: &instancepb.Instance{
			Config:      "emulator-config",
			DisplayName: "Test Instance",
			NodeCount:   1,
		},
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Println("Instance already exists, continuing...")
		} else {
			fmt.Fprintf(os.Stderr, "Failed to create instance: %v\n", err)
			os.Exit(1)
		}
	} else {
		if _, err := op.Wait(ctx); err != nil {
			if strings.Contains(err.Error(), "already exists") {
				fmt.Println("Instance already exists, continuing...")
			} else {
				fmt.Fprintf(os.Stderr, "Instance creation failed: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Println("Instance created successfully.")
		}
	}

	// Create database with DDL (or update DDL if database exists)
	fmt.Println("Creating database (if not exists)...")
	opDb, err := dbAdmin.CreateDatabase(ctx, &databasepb.CreateDatabaseRequest{
		Parent:          instanceName,
		CreateStatement: fmt.Sprintf("CREATE DATABASE `%s`", databaseID),
		ExtraStatements: []string{
			`CREATE TABLE products (
				product_id STRING(36) NOT NULL,
				name STRING(255) NOT NULL,
				description STRING(MAX),
				category STRING(100) NOT NULL,
				base_price_numerator INT64 NOT NULL,
				base_price_denominator INT64 NOT NULL,
				discount_percent INT64,
				discount_start_date TIMESTAMP,
				discount_end_date TIMESTAMP,
				status STRING(20) NOT NULL,
				created_at TIMESTAMP NOT NULL,
				updated_at TIMESTAMP NOT NULL,
				archived_at TIMESTAMP,
			) PRIMARY KEY (product_id)`,
			`CREATE INDEX idx_products_category ON products(category, status)`,
			`CREATE TABLE outbox_events (
				event_id STRING(36) NOT NULL,
				event_type STRING(100) NOT NULL,
				aggregate_id STRING(36) NOT NULL,
				payload JSON NOT NULL,
				status STRING(20) NOT NULL,
				created_at TIMESTAMP NOT NULL,
				processed_at TIMESTAMP,
			) PRIMARY KEY (event_id)`,
			`CREATE INDEX idx_outbox_status ON outbox_events(status, created_at)`,
		},
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Println("Database already exists. Schema may already be applied.")
			fmt.Printf("Database path: %s\n", databaseName)
			fmt.Println("\nSetup complete! You can now run: make run")
			return
		}
		fmt.Fprintf(os.Stderr, "Failed to create database: %v\n", err)
		os.Exit(1)
	}

	if _, err := opDb.Wait(ctx); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Println("Database already exists.")
		} else {
			fmt.Fprintf(os.Stderr, "Database creation failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("Database created successfully.")
	}

	fmt.Printf("\nDatabase ready: %s\n", databaseName)
	fmt.Println("\nSetup complete! You can now run: make run")
}
