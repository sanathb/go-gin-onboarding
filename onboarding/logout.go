package onboarding

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"

	"myMessages"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func Logout(db *sql.DB) gin.HandlerFunc {

	return func(c *gin.Context) {

		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered from panic")
				dump, _ := httputil.DumpRequest(c.Request, true)
				log.Printf("Pannic error caused by input %s ", dump)
				c.JSON(http.StatusInternalServerError, gin.H{"error": true, "Message": myMessages.PanicError})
			}
		}()

		token := c.Request.Header.Get("token")

		stmt, err := db.Prepare("DELETE FROM user_tokens WHERE token = ?")
		if err != nil {
			log.Fatal(err)
		}
		defer stmt.Close()

		res, err := stmt.Exec(token)

		if err != nil {
			log.Fatal(err)
		}

		rowCnt, err := res.RowsAffected()
		if err != nil {
			log.Fatal(err)
		}

		if rowCnt < 1 {
			//inavalid token
			var b BadRequest
			b.Error = true
			b.Message = myMessages.InvalidToken
			c.JSON(http.StatusBadRequest, b)
			return
		}

		type Output struct {
			Success bool
			Message string
		}

		var output Output
		output.Message = myMessages.LogoutSuccessful
		output.Success = true

		c.JSON(http.StatusOK, output)

	}
}
