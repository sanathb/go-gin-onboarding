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

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type BadRequest struct {
	Error   bool
	Message string
}

func Login(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered from panic")
				dump, _ := httputil.DumpRequest(c.Request, true)
				log.Printf("Pannic error caused by input %s ", dump)
				c.JSON(http.StatusInternalServerError, gin.H{"error": true, "Message": myMessages.PanicError})
			}
		}()

		// Binding from JSON
		type Login struct {
			Password string `json:"password" binding:"required"`
			Email    string `json:"email" binding:"required"`
		}

		var inputjson Login

		if c.BindJSON(&inputjson) != nil {

			var b BadRequest
			b.Error = true
			b.Message = myMessages.InvalidJson

			c.JSON(http.StatusBadRequest, b)
			return
		}

		password := myEncryption.GetSHA1(inputjson.Password + myConstants.PasswordSalt)

		transaction, _ := db.Begin()

		//Can't use Prepare statements if we use transactions
		rows, err := db.Query("SELECT id, email, fname, lname from pupils where email = '" + inputjson.Email + "' AND password = '" + password + "'")

		if err != nil {
			transaction.Rollback()
			log.Fatal(err)
		}
		defer rows.Close()

		var (
			id    int64
			email string
			fname string
			lname string
		)

		if rows.Next() != true {
			var b BadRequest
			b.Error = true
			b.Message = myMessages.InvalidUserCredential
			c.JSON(http.StatusOK, b)
			return
		} else {

			err := rows.Scan(&id, &email, &fname, &lname)
			if err != nil {
				transaction.Rollback()
				log.Fatal(err)
			}
		}

		//create a session token
		token := myRand.RandToken()
		res2, err := db.Exec("INSERT INTO user_tokens SET pupil_id = " + strconv.FormatInt(id, 10) + ", token = '" + token + "'")
		res2 = res2
		if err != nil {
			transaction.Rollback()
			log.Fatal(err)
		}

		transaction.Commit()

		type Output struct {
			Id      int64
			Fname   string
			Lname   string
			Email   string
			Message string
			Token   string
		}

		var output Output

		output.Id = id
		output.Email = email
		output.Fname = fname
		output.Lname = lname
		output.Message = myMessages.LoginSuccessful
		output.Token = token

		c.JSON(http.StatusOK, output)

	}
}
