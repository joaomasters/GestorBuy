// Package mongodb centraliza a conexão com o MongoDB e o bootstrap de
// índices. Toda a aplicação compartilha a mesma *mongo.Database resolvida
// aqui — nenhum outro pacote deve abrir sua própria conexão.
package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Connect abre a conexão com o cluster (local via Docker em dev, Atlas em
// produção — a URI é o único ponto de variação) e confirma com um Ping antes
// de devolver o handle, para falhar no boot em vez de na primeira requisição.
func Connect(ctx context.Context, uri, dbName string) (*mongo.Client, *mongo.Database, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, nil, fmt.Errorf("mongodb: falha ao conectar: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, nil, fmt.Errorf("mongodb: ping falhou: %w", err)
	}

	return client, client.Database(dbName), nil
}
