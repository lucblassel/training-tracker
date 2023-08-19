package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lucblassel/training-tracker/config"
	"github.com/lucblassel/training-tracker/controllers"
	"github.com/lucblassel/training-tracker/generator"
	"github.com/lucblassel/training-tracker/models"
)

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin": // MacOS
		cmd = "open"
	default: // various unix flavours
		cmd = "xdg-open"
	}

	args = append(args, url)

	return exec.Command(cmd, args...).Start()
}

//go:embed templates
var templateFiles embed.FS

//go:embed static
var staticFiles embed.FS

// func populate(c *gin.Context) {
// 	var runs []models.Run
//
// 	data, err := os.ReadFile("./runs.json")
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}
//
// 	json.Unmarshal(data, &runs)
//
// 	for _, run := range runs {
// 		log.Println(run.ID, run.Slug, run.CreatedAt, run.LastPulled)
// 		models.DB.Create(&run)
// 	}
//
// 	// models.DB.CreateInBatches(&runs, 100)
// }

func main() {
	if err := config.Init(); err != nil {
		log.Fatalf("Failed to configure app: %v", err)
	}

	log.Println(config.Config.AllSettings())
	config.Config.Debug()

	log.Println(config.Config.GetString("store.dir"))
	if err := os.MkdirAll(config.GetDataPath(), os.ModePerm); err != nil {
		log.Fatalf("Error initiallizing store dir: %v", err)
	}

	err := models.ConnectDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}

	if err := generator.InitGenerator(); err != nil {
		log.Fatalf("Failed to initialize random generator: %v", err)
	}

	router := gin.Default()

	templ := template.Must(template.New("").ParseFS(templateFiles, "templates/*.html"))
	router.SetHTMLTemplate(templ)
	// router.LoadHTMLGlob("./templates/*")

	// DEV TEST
	// router.GET("/populate", populate)
	//########

	router.GET("/", controllers.GetAllRuns)
	router.GET("/all", func(c *gin.Context) {
		c.Request.URL.Path = "/"
		router.HandleContext(c)
	})
	router.GET("/run/:slug", controllers.GetRun)
	router.GET("/pull/:slug", controllers.PullRun)
	router.GET("/edit/:slug", controllers.EditRun)
	router.GET("/duplicate/:slug", controllers.DuplicateRun)
	router.GET("/update", controllers.UpdateAll)
	router.GET("/new", controllers.CreateRun)

	router.POST("/toggle/:slug", controllers.ToggleRunFinished)
	router.POST("/fail/:slug", controllers.MarkRunFailed)
	router.POST("/success/:slug", controllers.MarkRunSuccessful)
	router.POST("/save/", controllers.SaveRun)
	router.POST("/delete/:slug", controllers.DeleteRun)

	newFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("Error loading static files: %v", err)
	}
	router.StaticFS("/files", http.FS(newFS))
	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.html", gin.H{"title": "Page not found"})
	})

	// Open Web Browser
	go func() {
		err := openBrowser("http://localhost:8080")
		if err != nil {
			log.Println("Could not open browser:", err)
		}
	}()

	// Launch Background Update task
	go func() {
		t := time.Tick(5 * time.Minute)
		for {
			select {
			case <-t:
				_, err := http.Get("http://localhost:8080/update")
				if err != nil {
					log.Println("Error Updating files: ", err)
				}
			}
		}
	}()

	if err := router.Run(); err != nil {
		log.Fatal(err)
	}
}
