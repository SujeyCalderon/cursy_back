package services

import (
	"context"
	"encoding/json"
	"fmt"
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
        log.Printf("❌ FIREBASE: No se pudo leer el archivo '%s': %v", jsonPath, err)
        log.Printf("❌ FIREBASE: Verifica que el archivo existe en el directorio de trabajo")
        return // ← IMPORTANTE: antes hacías initApp() aquí aunque falló la lectura
    }
    
    log.Printf("✅ FIREBASE: Archivo leído correctamente (%d bytes)", len(data))

    var creds map[string]interface{}
    if err := json.Unmarshal(data, &creds); err != nil {
        log.Printf("❌ FIREBASE: JSON inválido: %v", err)
        return
    }

    // Verificar campos requeridos
    projectID, _ := creds["project_id"].(string)
    clientEmail, _ := creds["client_email"].(string)
    tokenURI, _ := creds["token_uri"].(string)
    log.Printf("📋 FIREBASE: project_id=%s, client_email=%s, token_uri=%s", projectID, clientEmail, tokenURI)

    pk, ok := creds["private_key"].(string)
    if !ok || pk == "" {
        log.Printf("❌ FIREBASE: 'private_key' ausente o vacía en el JSON")
        return
    }

    cleanPK := strings.ReplaceAll(pk, "\\n", "\n")
    cleanPK = strings.ReplaceAll(cleanPK, "\\r", "")
    cleanPK = strings.TrimSpace(cleanPK)

    if !strings.HasPrefix(cleanPK, "-----BEGIN") {
        cleanPK = "-----BEGIN PRIVATE KEY-----\n" + cleanPK + "\n-----END PRIVATE KEY-----"
    }

    log.Printf("🔑 FIREBASE: Private key empieza con: %s", cleanPK[:50])

    creds["private_key"] = cleanPK
    cleanData, err := json.Marshal(creds)
    if err != nil {
        log.Printf("❌ FIREBASE: Error re-serializando JSON: %v", err)
        return
    }

    initApp(option.WithCredentialsJSON(cleanData))
}

func initApp(opt option.ClientOption) {
    config := &firebase.Config{ProjectID: "cursy-app"}
    app, err := firebase.NewApp(context.Background(), config, opt)
    if err != nil {
        log.Printf("❌ FIREBASE initApp: Fallo inicializar Firebase App: %v", err)
        return
    }
    log.Printf("✅ FIREBASE initApp: App creada correctamente")

    client, err := app.Messaging(context.Background())
    if err != nil {
        log.Printf("❌ FIREBASE initApp: Fallo obtener cliente FCM: %v", err)
        return
    }
    fcmClient = client
    log.Printf("✅ FIREBASE: FCM Client inicializado correctamente. fcmClient=%v", fcmClient != nil)
}


// SendPushNotification - VERSIÓN CORREGIDA
func SendPushNotification(token, title, body string, data map[string]string) error {
    if strings.TrimSpace(token) == "" {
        log.Println("ℹ️ Ignorando envío: token vacío")
        return nil
    }
    if fcmClient == nil {
        // ← Antes decía return nil, lo cual ocultaba el error
        log.Println("❌ CRÍTICO: FCM Client es nil — Firebase no se inicializó correctamente")
        return fmt.Errorf("FCM client no inicializado")
    }

    courseID := ""
    if data != nil {
        courseID = data["course_id"]
    }

    // ✅ SOLO DATA PAYLOAD — garantiza que onMessageReceived siempre se llame
    message := &messaging.Message{
        Token: token,
        Data: map[string]string{
            "title":     title,
            "body":      body,
            "type":      "new_course",
            "course_id": courseID,
        },
        Android: &messaging.AndroidConfig{
            Priority: "high",
            // Sin AndroidNotification — el cliente la construye
        },
        APNS: &messaging.APNSConfig{
            Headers: map[string]string{
                "apns-priority": "10",
            },
            Payload: &messaging.APNSPayload{
                Aps: &messaging.Aps{
                    ContentAvailable: true,
                    MutableContent:   true,
                    Sound:            "default",
                    Alert: &messaging.ApsAlert{
                        Title: title,
                        Body:  body,
                    },
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