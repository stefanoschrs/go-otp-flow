package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	bolt "go.etcd.io/bbolt"
)

/**
 * ENV:
 * @OTPF_ISSUER - stefanoschrs.com
 * @PORT - 8080
 */

const dbPath = "data.db"
const bucketName = "go-otp-flow"

var defaultIssuer string
var tmpKeys map[string]string
var db *bolt.DB

func getGenerate(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	issuer := c.DefaultQuery("issuer", defaultIssuer)

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: id,
	})
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	tmpKeys[fmt.Sprintf("%s:%s", issuer, id)] = key.Secret()

	base64Image, err := getBase64Image(key)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if c.Query("headless") != "true" {
		c.HTML(http.StatusOK, "/templates/generate.tmpl", gin.H{
			"qr": base64Image,
		})
		return
	}

	/**
	 * Headless
	 */
	if c.Query("type") == "image" {
		c.String(http.StatusOK, base64Image)
		return
	}

	c.String(http.StatusOK, key.String())
}

func postValidate(c *gin.Context) {
	var err error

	var body struct {
		Id string `json:"id" binding:"required"`
		Token string `json:"token" binding:"required"`
		Issuer string `json:"issuer"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	issuer := body.Issuer
	if issuer == "" {
		issuer = defaultIssuer
	}

	/**
	 * New client
	 */
	key := fmt.Sprintf("%s:%s", issuer, body.Id)
	if tmpKeys[key] != "" {
		valid := totp.Validate(body.Token, tmpKeys[key])
		if valid {
			err = db.Update(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte(bucketName))
				err := b.Put([]byte(key), []byte(tmpKeys[key]))
				return err
			})
			if err != nil {
				log.Println(err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			c.Status(http.StatusOK)
			return
		}

		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	/**
	 * Existing client
	 */
	var secret string
	_ = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		secret = string(b.Get([]byte(key)))
		return nil
	})
	if secret == "" {
		log.Println("Client not found")
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	valid := totp.Validate(body.Token, secret)
	if valid {
		c.Status(http.StatusOK)
	} else {
		c.AbortWithStatus(http.StatusForbidden)
	}
}

func main() {
	var err error

	defaultIssuer = os.Getenv("OTPF_ISSUER")
	if defaultIssuer == "" {
		defaultIssuer = "stefanoschrs.com"
	}

	/**
	 * Database
	 */
	db, err = bolt.Open(dbPath, 0666, nil)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b != nil {
			return nil
		}

		_, err := tx.CreateBucket([]byte(bucketName))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	tmpKeys = make(map[string]string)

	/**
	 * Server
	 */
	router := gin.Default()

	templates, err := loadTemplate()
	if err != nil {
		log.Fatal(err)
	}
	router.SetHTMLTemplate(templates)

	router.GET("/generate", getGenerate)
	router.POST("/validate", postValidate)

	if err = router.Run(); err != nil {
		log.Fatal(err)
	}
}
