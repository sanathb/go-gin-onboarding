package myMiddleware

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"

	"myConstants"
	"myMessages"

	"github.com/gin-gonic/gin"
	"github.com/xeipuuv/gojsonschema"
)

//Middleware for restricting input content
func RestrictInputContent(c *gin.Context) {
	fmt.Println(c.Request.ContentLength)
	if c.Request.ContentLength > myConstants.MaxInputContentLength {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"eror": true, "message": myMessages.ContentTooLarge})
		c.Abort()
		return
	}
}

//Middleware for checking token in header
func CheckToken(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		token := c.Request.Header.Get("token")

		//To do: change this to switch case style
		if strings.EqualFold(c.Request.RequestURI, "/signup") || strings.EqualFold(c.Request.RequestURI, "/login") {
			c.Next()
			return
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"eror": true, "message": myMessages.APITokenRequired})
			c.Abort()
			return
		}

		//check if the token exists in DB
		rows, err := db.Query("SELECT id FROM user_tokens WHERE token = '" + token + "'")

		if err != nil {
			log.Fatal(err)
		}

		defer rows.Close()
		if rows.Next() != true {
			c.JSON(http.StatusUnauthorized, gin.H{"error": true, "message": myMessages.InvalidToken})
			c.Abort()
			return
		}

		c.Next()
	}
}

//Middleware for logging requests
func RequestLoggerMiddleware() gin.HandlerFunc {
	fmt.Println("Request logger middleware initiated!")

	fo, err := os.Create(myConstants.LogFileLocation)
	if err != nil {
		fmt.Println(myMessages.FileCreationFailed + err.Error())
	} else {
		fmt.Println(myMessages.LogFileCreated + myConstants.LogFileLocation)
	}
	fo.Close()

	return func(c *gin.Context) {
		f, err := os.OpenFile(myConstants.LogFileLocation, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Println("Failed to open file" + err.Error())
		}
		defer f.Close()

		log.SetOutput(f)
		dump, err := httputil.DumpRequest(c.Request, true)

		log.Printf("Content type: "+c.ContentType()+" IP: "+
			c.ClientIP()+" header token: "+c.Request.Header.Get("token")+" Content length: "+strconv.FormatInt(c.Request.ContentLength, 10)+
			" Request Method: "+c.Request.Method+" Request URL: "+c.Request.RequestURI+"\nInput:\n%s", dump)

		c.Next()
	}
}

//middleware for json schema validation
func JsonSchemaValidator() gin.HandlerFunc {

	return func(c *gin.Context) {

		contextCopy := c.Copy()
		var schema string
		validateSchema := true

		switch c.Request.RequestURI {
		case "/signup":
			schema = "signup.json"
		case "/login":
			schema = "login.json"
		default:
			validateSchema = false
		}

		if validateSchema {

			schemaLoader := gojsonschema.NewReferenceLoader("file://" + myConstants.JsonSchemasFilesLocation + "/" +
				schema)

			body, _ := ioutil.ReadAll(contextCopy.Request.Body)
			//fmt.Printf("%s", body)

			documentLoader := gojsonschema.NewStringLoader(string(body))

			result, err := gojsonschema.Validate(schemaLoader, documentLoader)

			if err != nil {
				log.Fatalln(myMessages.JsonSchemaValidationError + err.Error())
			}

			if !result.Valid() {
				//fmt.Printf("The document is not valid. see errors :\n")
				type errorsList []string
				type JsonErrorOutput struct {
					Error   bool
					Message string
					Faults  errorsList
				}
				var el errorsList
				for _, desc := range result.Errors() {
					//fmt.Printf("- %s\n", desc)
					el = append(el, fmt.Sprintf("%s", desc))
				}
				var jeo JsonErrorOutput
				jeo.Error = true
				jeo.Message = myMessages.JsonSchemaValidationFailed
				jeo.Faults = el
				c.JSON(http.StatusBadRequest, jeo)
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
