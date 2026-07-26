package main

import (
	"embed"
	"log"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/OptLTD/member/internal/config"
	"github.com/OptLTD/member/internal/handle"
	"github.com/OptLTD/member/internal/routes"
	"github.com/OptLTD/member/internal/service"
)

//go:embed ui-front
var uiFront embed.FS

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	var cfg = config.Load()
	config.InitLogger(cfg)

	if err := config.InitDB(cfg); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	if err := service.HandleInit(); err != nil {
		log.Fatal("Failed to initialize data:", err)
	}

	gin.SetMode(cfg.AppMode)
	r := gin.Default()

	origins := []string{"*"}
	if v := os.Getenv("CORS_ORIGINS"); v != "" {
		origins = nil
		for _, o := range strings.Split(v, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				origins = append(origins, o)
			}
		}
	}
	allowCredentials := true
	for _, o := range origins {
		if o == "*" {
			allowCredentials = false
			break
		}
	}
	r.Use(cors.New(cors.Config{
		AllowHeaders:     []string{"*"},
		AllowOrigins:     origins,
		ExposeHeaders:    []string{"Content-Length"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: allowCredentials,
	}))

	r.Use(handle.StaticRouteHandle(cfg, uiFront))
	authService := service.NewAuthService()
	routes.Register(r, authService)

	log.Printf("Server starting on port %s", cfg.AppPort)
	log.Printf("Modules: %v", config.EnabledModules())
	log.Printf("Admin panel: http://localhost:%s/admin", cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
