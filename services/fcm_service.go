package services

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

var fcmClient *messaging.Client

func InitFirebase() {
	jsonPath := "firebase-service-account.json"
	data, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		log.Printf("Error leyendo credentials file: %v", err)
		opt := option.WithCredentialsFile(jsonPath)
		initApp(opt)
		return
	}

	// Limpieza de llave privada
	var creds map[string]interface{}
	if err := json.Unmarshal(data, &creds); err == nil {
		if pk, ok := creds["private_key"].(string); ok {
			cleanPK := strings.ReplaceAll(pk, "\\n", "\n")
			cleanPK = strings.ReplaceAll(cleanPK, "\\r", "")
			cleanPK = strings.Trim(cleanPK, " \t\n\r\"")
			
			if !strings.HasPrefix(cleanPK, "-----BEGIN") {
				cleanPK = "-----BEGIN PRIVATE KEY-----\n" + cleanPK + "\n-----END PRIVATE KEY-----"
			}
			creds["private_key"] = cleanPK
			cleanData, _ := json.Marshal(creds)
			initApp(option.WithCredentialsJSON(cleanData))
			return
		}
	}
	initApp(option.WithCredentialsJSON(data))
}

func initApp(opt option.ClientOption) {
	config := &firebase.Config{ProjectID: "cursy-app"}
	app, err := firebase.NewApp(context.Background(), config, opt)
	if err != nil {
		log.Printf("ERROR: Fallo inicializar Firebase App: %v", err)
		return
	}
	client, err := app.Messaging(context.Background())
	if err != nil {
		log.Printf("ERROR: Fallo obtener cliente FCM: %v", err)
		return
	}
	fcmClient = client
	log.Println("✅ Firebase Admin SDK inicializado")
}

// SendPushNotification - VERSIÓN CORREGIDA
func SendPushNotification(token, title, body string, data map[string]string) error {
	if strings.TrimSpace(token) == "" {
		log.Println("ℹ️ Ignorando envío: token vacío")
		return nil
	}
	if fcmClient == nil {
		log.Println("⚠️ FCM Client no inicializado")
		return nil
	}

	// Preparar data payload
	payload := make(map[string]string)
	if data != nil {
		for k, v := range data {
			payload[k] = v
		}
	}
	// Agregar metadata útil
	payload["click_action"] = "FLUTTER_NOTIFICATION_CLICK" // o tu propia acción
	payload["type"] = "new_course"

	tokenDisplay := token
	if len(tokenDisplay) > 10 {
		tokenDisplay = tokenDisplay[:10] + "..."
	}
	log.Printf("📱 Enviando a: %s", tokenDisplay)

	message := &messaging.Message{
		Token: token,
		// ✅ ESTO ES CLAVE: Notificación para background
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		// Datos para cuando la app está en foreground
		Data: payload,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				ChannelID: "new_courses_channel", // Debe coincidir con tu canal en Android
				Sound:     "default",
				// Para abrir app al tocar
				ClickAction: "OPEN_COURSE_DETAIL",
			},
		},
		// Opcional: para iOS
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
	}

	response, err := fcmClient.Send(context.Background(), message)
	if err != nil {
		log.Printf("❌ ERROR FCM: %v", err)
		return err
	}
	log.Printf("🚀 Notificación enviada. ID: %s", response)
	return nil
}

// SendMulticastNotification - Para enviar a múltiples tokens
func SendMulticastNotification(tokens []string, title, body string, data map[string]string) (*messaging.BatchResponse, error) {
	if fcmClient == nil {
		return nil, nil
	}

	// Filtrar tokens vacíos
	validTokens := make([]string, 0)
	for _, t := range tokens {
		if strings.TrimSpace(t) != "" {
			validTokens = append(validTokens, t)
		}
	}
	if len(validTokens) == 0 {
		return nil, nil
	}

	payload := make(map[string]string)
	if data != nil {
		for k, v := range data {
			payload[k] = v
		}
	}
	payload["type"] = "new_course"

	message := &messaging.MulticastMessage{
		Tokens: validTokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: payload,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				ChannelID: "new_courses_channel",
				Sound:     "default",
			},
		},
	}

	response, err := fcmClient.SendMulticast(context.Background(), message)
	if err != nil {
		log.Printf("❌ ERROR multicast: %v", err)
		return nil, err
	}

	log.Printf("🚀 Multicast: %d exitosos, %d fallidos", response.SuccessCount, response.FailureCount)
	return response, nil
}