package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"voyager/config"
	"voyager/pkg/logger"
	v "voyager/pkg/version"
	"voyager/router"
	"voyager/router/middleware"

	"github.com/YueHonghui/rfw"
	"github.com/gin-gonic/gin"

	//"github.com/lexkong/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	cfg     = pflag.StringP("config", "c", "", "voyager config file path.")
	version = pflag.BoolP("version", "v", false, "show version info.")
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	logrus.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	logrus.SetLevel(logrus.TraceLevel)

	// init config
	if err := config.Init(*cfg); err != nil {
		panic(err)
	}

	// Logging with rwf
	if err := os.MkdirAll(logger.LogDir, os.ModePerm); err != nil {
		logrus.WithError(err).Fatalf("Failed mkdir -p %s", logger.LogDir)
	}

	if rfw, err := rfw.NewWithOptions(filepath.Join(logger.LogDir, "access"), rfw.WithCleanUp(logger.LogRemain)); err == nil {
		//gin.DefaultWriter = rfw
		//gin.DefaultErrorWriter = rfw
		//gin.DisableConsoleColor()
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(rfw)

		logger.AccessLog = &logrus.Logger{
			Out:       rfw,
			Level:     logrus.InfoLevel,
			Formatter: &logrus.JSONFormatter{},
		}
		//defer rfw.Close()
	} else {
		logger.AccessLog = logrus.StandardLogger()
	}

	if rfw, err := rfw.NewWithOptions(filepath.Join(logger.LogDir, "runtime"), rfw.WithCleanUp(logger.LogRemain)); err == nil {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(rfw)

		logger.RuntimeLog = &logrus.Logger{
			Out:       rfw,
			Level:     logrus.DebugLevel,
			Formatter: &logrus.JSONFormatter{},
		}
		//defer rfw.Close()
	} else {
		logger.RuntimeLog = logrus.StandardLogger()
	}

	if rfw, err := rfw.NewWithOptions(filepath.Join(logger.LogDir, "metrics"), rfw.WithCleanUp(logger.LogRemain)); err == nil {
		logger.MetricsLog = &logrus.Logger{
			Out:       rfw,
			Level:     logrus.DebugLevel,
			Formatter: &logger.MetricsJSONFormatter{},
		}
		//defer rfw.Close()
	} else {
		logger.MetricsLog = logrus.StandardLogger()
	}
}

func main() {
	pflag.Parse()

	if *version {
		v := v.Get()
		marshalled, err := json.MarshalIndent(&v, "", "  ")
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(marshalled))
		return
	}

	// Set gin mode.
	gin.SetMode(viper.GetString("service.runmode"))

	// Create the Gin engine.
	g := gin.New()

	//middlewares := []gin.HandlerFunc{}

	// Routes.
	router.Load(
		// Cores.
		g,
		// Middlewares.
		middleware.Logging(),
		middleware.RequestId(),
	)

	// Ping the server to make sure the router is working.
	go func() {
		if err := pingServer(); err != nil {
			fmt.Println("The router has no response, or it might took too long to start up.\n", err)
		}
		fmt.Printf("The router has been deployed successfully.\n")
	}()

	fmt.Printf("Start to listening the incoming requests on http address: %s \n", viper.GetString("service.addr"))
	fmt.Printf(http.ListenAndServe(viper.GetString("service.addr"), g).Error())
}

// pingServer pings the http server to make sure the router is working.
func pingServer() error {

	for i := 0; i < viper.GetInt("service.maxPingCount"); i++ {
		// Ping the server by sending a GET request to `/health`.
		resp, err := http.Get(viper.GetString("service.url") + "/sd/health")
		if err == nil && resp.StatusCode == 200 {
			return nil
		}

		// Sleep for a second to continue the next ping.
		fmt.Printf("Waiting for the router, retry in 1 second.\n")
		time.Sleep(time.Second)
	}
	return errors.New("Cannot connect to the router.\n")
}
