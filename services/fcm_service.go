package services

import (
	"context"
	"log"
	"path/filepath"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

var fcmClient *messaging.Client

// InitFirebase inicializa el SDK de Firebase Admin
func InitFirebase() {
	serviceAccountKeyPath, _ := filepath.Abs("firebase-service-account.json")
	opt := option.WithCredentialsFile(serviceAccountKeyPath)
	
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Error inicializando Firebase App: %v", err)
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		log.Fatalf("Error obteniendo cliente de FCM: %v", err)
	}

	fcmClient = client
	log.Println("Firebase Admin SDK inicializado correctamente")
}

// SendPushNotification envía una notificación push a un token específico
func SendPushNotification(token, title, body string) error {
	if fcmClient == nil {
		log.Println("FCM Client no inicializado")
		return nil
	}

	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				ChannelID: "cursy_notifications",
				Priority:  messaging.PriorityHigh,
			},
		},
		Data: map[string]string{
			"title":        title,
			"body":         body,
			"click_action": "FLUTTER_NOTIFICATION_CLICK",
		},
	}

	response, err := fcmClient.Send(context.Background(), message)
	if err != nil {
		log.Printf("Error enviando notificación push: %v", err)
		return err
	}

	log.Printf("Notificación enviada con éxito: %s", response)
	return nil
}
