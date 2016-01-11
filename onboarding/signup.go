package onboarding

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"

	"myConstants"
	"myEncryption"
	"myMessages"
	"myRand"
	//_ "myRecovery"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func Signup(db *sql.DB) gin.HandlerFunc {

	return func(c *gin.Context) {

		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered from panic")
				dump, _ := httputil.DumpRequest(c.Request, true)
				log.Printf("Pannic error caused by input %s ", dump)
				c.JSON(http.StatusInternalServerError, gin.H{"error": true, "Message": myMessages.PanicError})
			}
		}()
		//myRecovery.Recover()
		//panic("I am a panic  error")

		// Binding from JSON
		type Login struct {
			Fname    string `json:"fname" binding:"required"`
			Lname    string `json:"lname"`
			Password string `json:"password" binding:"required"`
			Email    string `json:"email" binding:"required"`
		}

		var inputjson Login
		err := c.BindJSON(&inputjson)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": true, "Message": myMessages.InvalidJsonFormat})
			return
		}

		transaction, _ := db.Begin()

		//Can't use Prepare statements if we use transactions
		rows, err := db.Query("SELECT id, email from pupils where email = '" + inputjson.Email + "'")

		if err != nil {
			transaction.Rollback()
			log.Fatal(err)
		}
		defer rows.Close()

		if rows.Next() == true {
			c.JSON(http.StatusOK, gin.H{"error": true, "message": myMessages.UserEmailExists})
			return
		}

		password := myEncryption.GetSHA1(inputjson.Password + myConstants.PasswordSalt)

		res, err := db.Exec("INSERT INTO pupils SET fname = '" + inputjson.Fname + "', lname = '" + inputjson.Lname +
			"', email = '" + inputjson.Email + "', password = '" + password + "'")

		if err != nil {
			transaction.Rollback()
			log.Fatal(err)
		}

		lastId, err := res.LastInsertId()

		type Output struct {
			Id      int64
			Fname   string
			Lname   string
			Email   string
			Message string
			Token   string
		}

		var output Output

		output.Fname = inputjson.Fname
		output.Lname = inputjson.Lname
		output.Email = inputjson.Email
		output.Id = lastId
		output.Message = myMessages.UserCreated

		token := myRand.RandToken()

		output.Token = token

		//insert into tokens table
		_, err = db.Exec("INSERT INTO user_tokens SET pupil_id = " + strconv.FormatInt(lastId, 10) + ", token = '" + token + "'")
		if err != nil {
			transaction.Rollback()
			log.Fatal(err)
		}

		transaction.Commit()

		c.JSON(http.StatusOK, output)
		return
	}
}
