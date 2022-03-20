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
	Route   string `json:"route"`
	Execute string `json:"execute"`
	Query   string `json:"query"`
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

		if config.Services[i].Execute == "" {
			router.GET(config.Services[i].Route, getHandler(i))
		} else {
			router.POST(config.Services[i].Route, getHandler(i))
		}
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

func executeSql(sql string) {
	_, err := conn.Exec(context.Background(), sql)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Exec failed: %v\n", err)
		os.Exit(1)
	}
}

func getHandler(i int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.Services[i].Execute != "" {
			execStr := generateQueryStr(c, config.Services[i].Execute)
			executeSql(execStr)
		}

		queryStr := generateQueryStr(c, config.Services[i].Query)
		c.IndentedJSON(http.StatusOK, getRecords(queryStr))
	}
}

func generateQueryStr(c *gin.Context, queryTemplate string) string {
	jsonMap := GetJsonData(c)

	fmt.Println(jsonMap)

	re := regexp.MustCompile("{{(\\w|\\d|\\s|\\.)+}}") // find {{param}} tags in query string
	var tags = re.FindAllString(queryTemplate, -1)
	for i := 0; i < len(tags); i++ {
		tag := tags[i][2 : len(tags[i])-2]
		tagFields := strings.Split(tag, ".")
		if len(tagFields) < 2 {
			fmt.Fprintf(os.Stderr, "Invalid tag: "+tag)
			os.Exit(1)
		}
		if strings.ToLower(tagFields[0]) == "param" {
			param := c.Param(tagFields[1])
			queryTemplate = strings.Replace(queryTemplate, tags[i], param, -1)
		} else if strings.ToLower(tagFields[0]) == "body" {
			bodyItem := jsonMap[tagFields[1]].(string)
			queryTemplate = strings.Replace(queryTemplate, tags[i], bodyItem, -1)
		}

	}

	return queryTemplate
}

func GetJsonData(c *gin.Context) map[string]interface{} {
	data, _ := ioutil.ReadAll(c.Request.Body)

	jsonMap := make(map[string]interface{})

	if len(data) == 0 {
		return jsonMap
	}

	err := json.Unmarshal(data, &jsonMap)
	if err != nil {
		panic(err)
	}

	return jsonMap
}
