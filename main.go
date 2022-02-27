package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
)

type Config struct {
	DbConnString string    `json:"dbConnString"`
	Services     []Service `json:"services"`
}

type Service struct {
	Route string `json:"route"`
	Query string `json:"query"`
}

var conn *pgx.Conn
var config Config

func main() {

	// 1. Read Config File
	jsonFile, err := os.Open("config.json")

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer jsonFile.Close()

	fmt.Println("Successfully opened config.json")

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal(byteValue, &config)

	// 2. Open DB connection
	openDB(config.DbConnString)
	defer closeDB()

	// 3. Create services
	router := gin.Default()

	for i := 0; i < len(config.Services); i++ {
		fmt.Println("Route : " + config.Services[i].Route + " Query : " + config.Services[i].Query)
		router.GET(config.Services[i].Route, getHandler(i))

	}

	// 4. Run services
	router.Run("localhost:8080")
}

func openDB(connString string) {
	var err error
	conn, err = pgx.Connect(context.Background(), connString)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
}

func closeDB() {
	defer conn.Close(context.Background())
}

func getRecords(query string) interface{} {
	rows, err := conn.Query(context.Background(), query)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
		os.Exit(1)
	}

	defer rows.Close()

	result := make([]interface{}, 0)

	for rows.Next() {
		values, _ := rows.Values()
		fieldDescriptions := rows.FieldDescriptions()

		fieldValueMap := make(map[string]interface{}, len(values))

		for j := 0; j < len(values); j++ {
			fieldValueMap[string(fieldDescriptions[j].Name)] = values[j]
		}

		result = append(result, fieldValueMap)
	}

	if len(result) == 1 {
		return result[0]
	} else {
		return result
	}

}

func getHandler(i int) gin.HandlerFunc {
	return func(c *gin.Context) {
		queryStr := generateQueryStr(c, config.Services[i].Query)
		c.IndentedJSON(http.StatusOK, getRecords(queryStr))
	}
}

func generateQueryStr(c *gin.Context, queryTemplate string) string {
	re := regexp.MustCompile("{{(\\w|\\d|\\s)+}}") // find {{param}} tags in query string
	var tags = re.FindAllString(queryTemplate, -1)
	for i := 0; i < len(tags); i++ {
		tag := tags[0][2 : len(tags[0])-2]
		param := c.Param(tag)
		queryTemplate = strings.Replace(queryTemplate, tags[0], param, -1)
	}

	return queryTemplate
}
