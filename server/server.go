package server

import (
	"fmt"
	"log"
	"os"
	"studious-waffle/server/protodata"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
)

var baseUrl = "http://localhost:"

func ServeGin() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	seed := os.Getenv("SEED_DATA")
	if seed == "" {
		log.Println("WARNING: SEED_DATA is not set.")
	}

	if seed == "true" {
		var wg sync.WaitGroup
		wg.Add(5)

		go func() {
			fmt.Println("Generating Routes...")
			haveData := protodata.GenerateRouteData()
			if haveData {
				fmt.Println("Finished generating Routes data.")
			}
			wg.Done()
		}()

		go func() {
			fmt.Println("Generating Trips...")
			haveData := protodata.GenerateTripData()
			if haveData {
				fmt.Println("Finished generating Trips data.")
			}
			wg.Done()
		}()

		go func() {
			fmt.Println("Generating Stops...")
			haveData := protodata.GenerateStopData()
			if haveData {
				fmt.Println("Finished generating Stops data.")
			}
			wg.Done()
		}()

		go func() {
			fmt.Println("Generating Stop Times...")
			haveData := protodata.GenerateStopTimeData()
			if haveData {
				fmt.Println("Finished generating Stop Times data.")
			}
			wg.Done()
		}()

		go func() {
			fmt.Println("Generating Shapes...")
			haveData := protodata.GenerateShapeData()
			if haveData {
				fmt.Println("Finished generating Shapes data.")
			}
			wg.Done()
		}()

		wg.Wait()
		fmt.Println("Server seeded with data.")
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.SetTrustedProxies(nil)

	r.Use(logger.SetLogger())
	r.Use(RateLimiter())
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:8080"}
	r.Use(cors.New(config))

	addBaseRoutes(r)
	AddGTFSRoutes(r)

	log.Printf("Serving Gin at %s\n", baseUrl+port)
	r.Run()
}

func addBaseRoutes(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World!",
		})
	})
}
