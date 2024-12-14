package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection

func main() {
	// Configurar cliente de MongoDB
	client, err := mongo.NewClient(options.Client().ApplyURI(getEnv("MONGO_URI", "mongodb://localhost:27017")))
	if err != nil {
		log.Fatalf("Error creando cliente de MongoDB: %v", err)
	}

	// Conectar al servidor
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Error conectando a MongoDB: %v", err)
	}

	// Verificar conexión
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("No se pudo conectar a MongoDB: %v", err)
	}

	fmt.Println("Conexión exitosa a MongoDB")

	// Configurar colección
	collection = client.Database("testdb").Collection("users")

	// Configurar servidor HTTP con Gin
	r := gin.Default()

	// Endpoint para probar la API
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "API funcionando"})
	})

	// Endpoint para obtener usuarios
	r.GET("/users", getUsers)

	// Iniciar servidor en puerto configurable
	port := getEnv("PORT", "8080")
	fmt.Printf("Servidor corriendo en el puerto %s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Error iniciando el servidor: %v", err)
	}
}

// getUsers: Ejemplo de endpoint para obtener usuarios
func getUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var users []bson.M
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener usuarios"})
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var user bson.M
		if err := cursor.Decode(&user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al leer usuario"})
			return
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, users)
}

// getEnv: Obtiene una variable de entorno con valor predeterminado
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
