package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type RequestBody struct {
	Title    string `json:"title"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
}

type Result struct {
	ID    int64  `db:"idinfo" json:"id"`
	Hash  string `db:"hashinfo" json:"hash"`
	Title string `db:"titleinfo" json:"title"`
	Date  string `db:"dateinfo" json:"date"`
	Size  int64  `db:"sizeinfo" json:"size"`
}

func main() {
	db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/rarbg")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	router.POST("/search", func(c *gin.Context) {
		var requestBody RequestBody

		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		titleParts := strings.Fields(requestBody.Title)

		var likeFilters []string
		for _, part := range titleParts {
			likeFilters = append(likeFilters, "titleinfo LIKE '%"+part+"%'")
		}

		whereClause := strings.Join(likeFilters, " AND ")

		page := requestBody.Page
		pageSize := requestBody.PageSize
		offset := (page - 1) * pageSize

		rows, err := db.QueryContext(
			context.TODO(),
			"SELECT * FROM info WHERE "+whereClause+" LIMIT ?, ?",
			offset, pageSize,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var results []Result
		for rows.Next() {
			var result Result
			if err := rows.Scan(
				&result.ID,
				&result.Hash,
				&result.Title,
				&result.Date,
				&result.Size,
			); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			results = append(results, result)
		}

		c.JSON(http.StatusOK, results)
	})

	err = router.RunTLS(":11001", "/etc/letsencrypt/live/noiduser.site/fullchain.pem", "/etc/letsencrypt/live/noiduser.site/privkey.pem")
	if err != nil {
		log.Fatal(err)
	}
}
