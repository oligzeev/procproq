package service

import (
	"context"
	"encoding/json"
	"example.com/oligzeev/pp-gin/internal/domain"
	log "github.com/sirupsen/logrus"
	"testing"
)

const (
	mappingStr1 = `{
  "id": "3028f11a-46c2-4739-b9c0-fa4024c0f7b3",
  "body": {
    "key1": "$.id"
  }
}`
	mappingStr4 = `{
  "id": "3028f11a-46c2-4739-b9c0-fa4024c0f7b3",
  "body": {
    "key1": "$.id",
    "key2": "$.product.id",
	"key3": "$.id",
    "key4": "$.product.id"
  }
}`
	mappingStr8 = `{
  "id": "3028f11a-46c2-4739-b9c0-fa4024c0f7b3",
  "body": {
    "key1": "$.id",
    "key2": "$.product.id",
	"key3": "$.id",
    "key4": "$.product.id",
    "key5": "$.id",
    "key6": "$.product.id",
	"key7": "$.id",
    "key8": "$.product.id"
  }
}`
	mappingStr16 = `{
  "id": "3028f11a-46c2-4739-b9c0-fa4024c0f7b3",
  "body": {
    "key1": "$.id",
    "key2": "$.product.id",
	"key3": "$.id",
    "key4": "$.product.id",
    "key5": "$.id",
    "key6": "$.product.id",
	"key7": "$.id",
    "key8": "$.product.id",
    "key9": "$.id",
    "key10": "$.product.id",
	"key11": "$.id",
    "key12": "$.product.id",
    "key13": "$.id",
    "key14": "$.product.id",
	"key15": "$.id",
    "key16": "$.product.id"
  }
}`

	bodyStr1 = `{
  "id": "111",
  "product": {
    "id": "222",
    "specification": {
     "id": "333"
    }
  }
}`
)

// go test -bench='BenchmarkBuildStartJobBody1$' -benchtime=5s -cpuprofile=cpu.out
// go tool pprof -http :8090 service.test cpu.out
func BenchmarkBuildStartJobBody1(b *testing.B) {
	var err error
	ctx := context.Background()
	mapping, body := unmarshalTD(mappingStr1, bodyStr1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err = buildStartJobBody(ctx, mapping, body); err != nil {
			log.Fatalf("can't make test: %v", err)
		}
	}
}

func BenchmarkBuildStartJobBody4(b *testing.B) {
	var err error
	mapping, body := unmarshalTD(mappingStr4, bodyStr1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err = buildStartJobBody(context.Background(), mapping, body); err != nil {
			log.Fatalf("can't make test: %v", err)
		}
	}
}

func BenchmarkBuildStartJobBody8(b *testing.B) {
	var err error
	mapping, body := unmarshalTD(mappingStr8, bodyStr1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err = buildStartJobBody(context.Background(), mapping, body); err != nil {
			log.Fatalf("can't make test: %v", err)
		}
	}
}

func BenchmarkBuildStartJobBody16(b *testing.B) {
	var err error
	mapping, body := unmarshalTD(mappingStr16, bodyStr1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err = buildStartJobBody(context.Background(), mapping, body); err != nil {
			log.Fatalf("can't make test: %v", err)
		}
	}
}

func unmarshalTD(mappingStr, bodyStr string) (*domain.ReadMapping, domain.Body) {
	var mapping domain.ReadMapping
	if err := json.Unmarshal([]byte(mappingStr), &mapping); err != nil {
		log.Fatalf("can't unmarshal read mapping")
	}
	var body domain.Body
	if err := json.Unmarshal([]byte(bodyStr), &body); err != nil {
		log.Fatalf("can't unmarshal body")
	}
	return &mapping, body
}
