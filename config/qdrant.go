package config

import (
	"context"
	"fmt"
	"log"

	"github.com/qdrant/go-client/qdrant"
)

var QdrantClient *qdrant.Client

func ConnectQdrant() {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "localhost",
		Port: 6334,
	})
	if err != nil {
		log.Fatal("Qdrant connection failed:", err)
	}

	QdrantClient = client
	fmt.Println("Qdrant connected successfully")
}

func CreateQdrantCollection() {
	ctx := context.Background()

	// check if collection already exists
	exists, err := QdrantClient.CollectionExists(ctx, "content_vectors")
	if err != nil {
		log.Fatal("Failed to check collection:", err)
	}

	if exists {
		fmt.Println("Qdrant collection already exists")
		return
	}

	// create collection with 768 dimensions (nomic-embed-text output size)
	err = QdrantClient.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: "content_vectors",
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     768,
			Distance: qdrant.Distance_Cosine,
		}),
	})
	if err != nil {
		log.Fatal("Failed to create collection:", err)
	}

	fmt.Println("Qdrant collection created successfully")
}
