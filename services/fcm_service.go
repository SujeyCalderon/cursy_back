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

// InitFirebase inicializa el SDK de Firebase Admin con limpieza profunda de llave privada
func InitFirebase() {
	jsonPath := "firebase-service-account.json"
	data, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		log.Printf("Error leyendo credentials file: %v. Intentando con lo que haya...", err)
		opt := option.WithCredentialsFile(jsonPath)
		initApp(opt)
		return
	}

	// Limpieza ULTRA ROBUSTA de la llave privada
	var creds map[string]interface{}
	if err := json.Unmarshal(data, &creds); err == nil {
		if pk, ok := creds["private_key"].(string); ok {
			log.Printf("INICIANDO LIMPIEZA DE LLAVE (Len: %d)", len(pk))
			
			// Paso 1: Reemplazar escapes literales comunes
			cleanPK := strings.ReplaceAll(pk, "\\n", "\n")
			cleanPK = strings.ReplaceAll(cleanPK, "\\r", "")
			
			// Paso 2: Limpiar posibles comillas dobles escapadas o basura
			cleanPK = strings.Trim(cleanPK, " \t\n\r\"")
			
			// Paso 3: Si la llave no tiene los headers, algo anda muy mal
			if !strings.Contains(cleanPK, "BEGIN PRIVATE KEY") {
				log.Println("ADVERTENCIA: La llave no parece tener el formato PEM correcto")
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
		log.Fatalf("CRÍTICO: Fallo al inicializar Firebase App: %v", err)
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		log.Fatalf("CRÍTICO: Fallo al obtener cliente FCM: %v", err)
	}

	fcmClient = client
	log.Println("✅ [ULTRA ROBUSTO] Firebase Admin SDK inicializado exitosamente")
}

// SendPushNotification envía una notificación push a un token específico con datos adicionales opcionales
func SendPushNotification(token, title, body string, data map[string]string) error {
	if fcmClient == nil {
		log.Println("FCM Client no inicializado")
		return nil
	}

	// Combinar el título y cuerpo base con los datos proporcionados
	payload := map[string]string{
		"title": title,
		"body":  body,
	}
	for k, v := range data {
		payload[k] = v
	}

	message := &messaging.Message{
		Token: token,
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		Data: payload,
	}

	response, err := fcmClient.Send(context.Background(), message)
	if err != nil {
		log.Printf("Error enviando notificación push: %v", err)
		return err
	}

	log.Printf("Notificación push enviada con éxito (%s): %s", title, response)
	return nil
}
