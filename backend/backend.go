package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
)

type pesanResponse struct {
	Message string `json:"message"`
}

type Tiket struct {
	NamaDepan    string `json:"namaDepan"`
	NamaBelakang string `json:"namaBelakang"`
	Email        string `json:"email"`
	Tiket        string `json:"tiket"`
}

func membuatTiket(c *gin.Context) {
	var requestBody Tiket
	if err := c.BindJSON(&requestBody); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "request payload salah"})
		return
	}

	namaDepan := requestBody.NamaDepan
	namaBelakang := requestBody.NamaBelakang
	email := requestBody.Email
	tiket := requestBody.Tiket
	if validasiInputData(namaDepan, namaBelakang, email, tiket) {
		// notifikasi jika Kafka down
		producer, err := newKafkaProducer()
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "server kafka down, silahkan cek server kafka nya"})
			return
		}
		defer producer.Close()

		// kirim pesan ke Kafka
		message := fmt.Sprintf("%v;%v;%v;%v", namaDepan, namaBelakang, email, tiket)
		message = fmt.Sprintf("Terima kasih %v %v telah memesan %v tiket untuk acara Jakarta Gigs Bawah Tanah. Silahkan cek email %v untuk konfirmasi selanjutnya.", namaDepan, namaBelakang, tiket, email)
		msg := &sarama.ProducerMessage{
			Topic: "tiket-topic",
			Value: sarama.StringEncoder(message),
		}
		_, _, err = producer.SendMessage(msg)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "pesan yang di kirim ke kafka gagal"})
			return
		}
		response := pesanResponse{Message: message}
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data yang di input salah"})
		return
	}
}

func validasiInputData(namaDepan string, namaBelakang string, email string, tiket string) bool {
	if strings.TrimSpace(namaDepan) == "" || strings.TrimSpace(namaBelakang) == "" || strings.TrimSpace(email) == "" || strings.TrimSpace(tiket) == "" {
		return false
	}
	return true
}

func main() {
	r := gin.Default()

	r.Use(CORSMiddleware())

	r.POST("/tiket", membuatTiket)

	if err := r.Run(":8080"); err != nil {
		log.Fatal("Error saat memulai backend server: ", err)
	}
}

// mengizinkan cors "*"
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "DELETE, POST, GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func newKafkaProducer() (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	// konfigurasi kafka
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	brokers := []string{"kafka:9092"}

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}
	return producer, nil
}
