package main

import (
	"go-integral/internal/seedgen"
	"log"
	"os"
)

func main() {
	sqlFilePath := "schema.sql"

	sqlContents, err := os.ReadFile(sqlFilePath)
	if err != nil {
		log.Fatalf("failed to read schema.sql: %v", err)
	}

	builder, err := seedgen.NewFromSQLSchema(string(sqlContents))
	if err != nil {
		log.Fatalf("failed to create builder: %v", err)
	}

	files, err := builder.GenerateTableSchemas()
	if err != nil {
		log.Fatalf("failed to generate table schemas: %v", err)
	}

	generatedFolder := "generated/seed"
	os.MkdirAll(generatedFolder, 0755)
	for _, file := range files {
		os.WriteFile(generatedFolder+"/"+file.Filename, []byte(file.Contents), 0644)
	}
}
