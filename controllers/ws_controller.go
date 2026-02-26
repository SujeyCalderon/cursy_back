package controllers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"cursy_back/config"
	"cursy_back/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Upgrader para convertir HTTP a WebSocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client representa al usuario conectado mediante WebSocket
type Client struct {
	ID   string
	Conn *websocket.Conn
	Send chan []byte
}

// Hub gestiona todas las conexiones activas
type Hub struct {
	Clients    map[string]*Client
	Register   chan *Client
	Unregister chan *Client
	Mutex      sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[string]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

var MainHub = NewHub()

// Run inicia el bucle del Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Mutex.Lock()
			// Si el usuario ya tiene una conexión activa, cerrarla primero
			if oldClient, exists := h.Clients[client.ID]; exists {
				close(oldClient.Send)
				oldClient.Conn.Close()
			}
			h.Clients[client.ID] = client
			h.Mutex.Unlock()
			log.Printf("Usuario conectado: %s", client.ID)

			// Primero enviamos al nuevo usuario el estado de los que ya están conectados (Snapshot)
			h.sendInitialStatusSnapshot(client)

			// Luego avisamos a los demás que este usuario se conectó
			h.broadcastStatus(client.ID, "online")

		case client := <-h.Unregister:
			h.Mutex.Lock()
			if _, ok := h.Clients[client.ID]; ok {
				delete(h.Clients, client.ID)
				close(client.Send)
				h.Mutex.Unlock() // Desbloquear antes de broadcast
				log.Printf("Usuario desconectado: %s", client.ID)
				h.broadcastStatus(client.ID, "offline")
			} else {
				h.Mutex.Unlock()
			}
		}
	}
}

// sendInitialStatusSnapshot envía al cliente recién conectado el estado de todos los demás
func (h *Hub) sendInitialStatusSnapshot(newClient *Client) {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	for id := range h.Clients {
		if id == newClient.ID {
			continue
		}

		message := map[string]string{
			"type":      "user_status",
			"sender_id": id,
			"content":   "online",
		}
		data, _ := json.Marshal(message)

		select {
		case newClient.Send <- data:
		default:
			// Canal lleno
		}
	}
}

// broadcastStatus envía un evento de estatus a todos los usuarios conectados
func (h *Hub) broadcastStatus(userID string, status string) {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	message := map[string]string{
		"type":      "user_status",
		"sender_id": userID,
		"content":   status,
	}
	data, _ := json.Marshal(message)

	for id, client := range h.Clients {
		if id != userID { // No enviárselo a uno mismo
			select {
			case client.Send <- data:
			default:
				// Si el canal está lleno, desconectar al cliente (o ignorar)
			}
		}
	}
}

func (c *Client) ReadPump() {
	defer func() {
		MainHub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, messageData, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		// El mensaje que llega del cliente debe tener: conversation_id, receiver_id y text
		var input struct {
			Type           string `json:"type"`
			ConversationID string `json:"conversation_id"`
			ReceiverID     string `json:"receiver_id"`
			Content        string `json:"content"`
			SenderID       string `json:"sender_id"` // Campo opcional al llegar, lo llenaremos nosotros
		}

		if err := json.Unmarshal(messageData, &input); err != nil {
			log.Printf("Error al decodificar mensaje: %v", err)
			continue
		}

		// Por defecto el tipo es chat si viene vacío
		if input.Type == "" {
			input.Type = "chat"
		}

		// Llenamos el sender_id con el ID del cliente que tiene el socket abierto
		input.SenderID = c.ID

		// Guardar el mensaje en la base de datos
		saveChatMessage(c.ID, input.ConversationID, input.Content)

		// Reenviar el mensaje al destinatario si está conectado
		// Volvemos a serializar para incluir el sender_id
		forwardData, _ := json.Marshal(input)

		MainHub.Mutex.Lock()
		if receiverClient, ok := MainHub.Clients[input.ReceiverID]; ok {
			receiverClient.Send <- forwardData
		}
		MainHub.Mutex.Unlock()
	}
}

func (c *Client) WritePump() {
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.Conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

// WSHandler maneja la conexión WebSocket
func WSHandler(c *gin.Context) {
	// Obtener ID del usuario del token JWT (el middleware ya lo puso en el contexto)
	userIDStr, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No autorizado"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Error al subir a WebSocket: %v", err)
		return
	}

	client := &Client{
		ID:   userIDStr.(string),
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	MainHub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}

func saveChatMessage(senderIDStr, conversationIDStr, content string) {
	senderID, _ := primitive.ObjectIDFromHex(senderIDStr)
	conversationID, _ := primitive.ObjectIDFromHex(conversationIDStr)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//Guardar mensaje
	msgCollection := config.GetCollection("messages")
	msg := models.Message{
		ID:             primitive.NewObjectID(),
		ConversationID: conversationID,
		SenderID:       senderID,
		Content:        content,
		CreatedAt:      time.Now(),
	}
	msgCollection.InsertOne(ctx, msg)

	//Actualizar conversación con el último mensaje
	convCollection := config.GetCollection("conversations")
	convCollection.UpdateOne(ctx, bson.M{"_id": conversationID}, bson.M{
		"$set": bson.M{
			"last_message": content,
			"updated_at":   time.Now(),
		},
	})
}
