// Package main is the entry point for the product catalog service
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	pb "product-catalog-service/proto/product/v1"

	"product-catalog-service/internal/services"
)

func main() {
	ctx := context.Background()

	// Get configuration from environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}

	spannerDatabase := os.Getenv("SPANNER_DATABASE")
	if spannerDatabase == "" {
		// Default to emulator database
		spannerDatabase = "projects/test-project/instances/test-instance/databases/test-database"
	}

	// Set emulator host if using emulator
	emulatorHost := os.Getenv("SPANNER_EMULATOR_HOST")
	if emulatorHost == "" {
		os.Setenv("SPANNER_EMULATOR_HOST", "localhost:9010")
	}

	log.Printf("Starting Product Catalog Service on port %s", port)
	log.Printf("Connecting to Spanner database: %s", spannerDatabase)

	// Initialize dependency container
	container, err := services.NewContainer(ctx, spannerDatabase)
	if err != nil {
		log.Fatalf("Failed to initialize service container: %v", err)
	}
	defer container.Close()

	log.Println("Service dependencies initialized successfully")

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register product service
	pb.RegisterProductServiceServer(grpcServer, container.ProductHandler)

	// Register reflection service (useful for grpcurl and testing)
	reflection.Register(grpcServer)

	// Create listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	log.Printf("gRPC server listening on :%s", port)

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh

		log.Println("Shutting down gracefully...")
		grpcServer.GracefulStop()
		log.Println("Server stopped")
	}()

	// Start serving
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
