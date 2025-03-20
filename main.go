package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Layer-Edge/light-node/node"
	"github.com/Layer-Edge/light-node/utils"
	"github.com/joho/godotenv"
)

// Worker is responsible for running the sample collection and verification process
func Worker(ctx context.Context, wg *sync.WaitGroup, id int) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Worker %d is shutting down\n", id)
			return
		default:
			fmt.Printf("Worker %d is running...\n", id)
			node.CollectSampleAndVerify()
			time.Sleep(5 * time.Second)
		}
	}
}

func main() {
	// Load environment variables from .env
	err := godotenv.Load("/root/list-node/light-node/.env") // Use absolute path
	if err != nil {
		log.Fatalf("❌ Error loading .env file: %v", err)
	}

	fmt.Println("✅ ENV Loaded Successfully")

	// Try reading environment variables for debugging
	grpcURL := os.Getenv("GRPC_URL")
	if grpcURL == "" {
		log.Fatal("❌ GRPC_URL is not set in .env file")
	}
	fmt.Printf("🔹 GRPC_URL: %s\n", grpcURL)

	privateKey := os.Getenv("PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal("❌ PRIVATE_KEY is not set in .env file")
	}
	fmt.Println("🔹 PRIVATE_KEY loaded successfully")

	// Get Public Key from utils
	pubKey, err := utils.GetCompressedPublicKey()
	if err != nil {
		log.Fatalf("❌ Failed to get compressed public key: %v", err)
	}
	fmt.Printf("🔹 Compressed Public Key: %s\n", pubKey)

	// Initialize context for handling graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	signalChan := make(chan os.Signal, 1)

	// Capture system signals for a clean shutdown
	signal.Notify(signalChan, syscall.SIGABRT, syscall.SIGTERM, syscall.SIGINT)

	wg.Add(1)
	go Worker(ctx, &wg, 1)

	// Wait for interrupt signal
	<-signalChan
	fmt.Println("\n⚠️ Received interrupt signal. Shutting down gracefully...")

	// Cancel worker process
	cancel()
	wg.Wait()
	fmt.Println("✅ Worker has shut down. Exiting..")
}
