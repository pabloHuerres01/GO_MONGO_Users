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
	"go.mongodb.org/mongo-driver/bson/primitive"
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

	// Endpoint para eliminar usuarios
	r.DELETE("/users/:id", deleteUser)

	// Endpoint para ingresar usuarios
	r.POST("/users", createUser)

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

// deleteUser: Función para eliminar un usuario
func deleteUser(c *gin.Context) {
	id := c.Param("id") // Obtener el ID del usuario desde la URL

	// Convertir el ID a ObjectID (si usas MongoDB con ObjectID)
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	// Intentar eliminar el usuario
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al eliminar usuario"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usuario eliminado"})
}

// getEnv: Obtiene una variable de entorno con valor predeterminado
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func createUser(c *gin.Context) {
	// Estructura para recibir datos del cliente
	var user struct {
		Name      string    `json:"name" binding:"required"`
		Email     string    `json:"email" binding:"required,email"`
		Age       int       `json:"age" binding:"required"`
		CreatedAt time.Time `json:"createdAt"`
	}

	// Validar datos del cuerpo de la solicitud
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Agregar la fecha de creación
	user.CreatedAt = time.Now()

	// Insertar en la base de datos
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al insertar usuario"})
		return
	}

	// Responder con el ID del usuario creado
	c.JSON(http.StatusOK, gin.H{
		"message": "Usuario creado",
		"id":      result.InsertedID,
	})
}
