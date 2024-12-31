package main

import (
	"github.com/gin-gonic/gin"
	"github.com/max-chem-eng/go-more-remote/controllers"
	"github.com/max-chem-eng/go-more-remote/models"
)

func main() {
	models.CreateConnection()
	models.CreateTables()

	r := gin.Default()
	controllers.SetupRoutes(r)

	r.Run(":8080")
}

// func main() {
// 	scriptContent := `
// 		require 'json'

// 		json_data = '{"name": "Max", "age": 30}'
// 		parsed = JSON.parse(json_data)
// 		puts "Name: #{parsed['name']}, Age: #{parsed['age']}"
// 	`
// 	language := "ruby"

// 	scriptPath, err := jobengine.SaveScriptToFile(scriptContent, language)
// 	if err != nil {
// 		fmt.Printf("Error saving script: %v\n", err)
// 		os.Exit(1)
// 	}
// 	defer os.Remove(scriptPath)

// 	logs, err := jobengine.ExecuteJob(jobengine.JobConfig{
// 		Language:   language,
// 		ScriptPath: scriptPath,
// 	})
// 	if err != nil {
// 		fmt.Printf("Error executing job: %v\n", err)
// 		os.Exit(1)
// 	}

// 	fmt.Printf("Job Output:\n%s\n", logs)
// }
