package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"github.com/google/uuid"
	"github.com/ma111e/excuses/internal/types"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strings"
	"time"
)

var (
	baseURL   = "https://cyber.excusesecu.fr/"
	log       = logrus.New()
	port      string
	debugMode bool
	cfgFile   string
)

type QuoteServer struct{}

func init() {
	cobra.OnInitialize(initConfig)

	// Configure logrus
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		PadLevelText:    true,
	})

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatal("Could not create logs directory:", err)
	}

	// Open log file
	file, err := os.OpenFile("logs/server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Could not open log file:", err)
	}

	// Set log output to both file and stdout
	//log.SetOutput(io.MultiWriter(os.Stdout, file))
	mw := io.MultiWriter(os.Stdout, file)
	log.SetOutput(mw)

	log.SetLevel(logrus.InfoLevel)
}

func initConfig() {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	viper.SetConfigName("excuses-server")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	viper.AddConfigPath(home)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Warn("No config file found, using defaults")
	}

	baseURL = viper.GetString("baseURL")
	if baseURL == "" {
		baseURL = "https://cyber.excusesecu.fr/"
	}
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "server",
		Short: "Start the quote server",
		Run: func(cmd *cobra.Command, args []string) {
			port = viper.GetString("port")
			debugMode = viper.GetBool("debug")

			if debugMode {
				log.SetLevel(logrus.DebugLevel)
				log.Debug("Debug logging enabled")
			}

			server := new(QuoteServer)
			_ = rpc.Register(server)
			rpc.HandleHTTP()

			listener, err := net.Listen("tcp", ":"+port)
			if err != nil {
				log.WithFields(logrus.Fields{"port": port, "error": err}).Fatal("Failed to start listener")
			}

			log.WithFields(logrus.Fields{"port": port}).Info("Server starting")
			go collectMetrics()

			if err := http.Serve(listener, nil); err != nil {
				log.WithFields(logrus.Fields{"error": err}).Fatal("Server failed")
			}
		},
	}

	rootCmd.PersistentFlags().StringVarP(&port, "port", "p", "1234", "Port to listen on")
	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug", "d", false, "Enable debug logging")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./excuses-client.yml)")

	_ = viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (s *QuoteServer) FetchQuote(req *types.FetchQuoteRequest, resp *types.FetchQuoteResponse) error {
	requestID := generateRequestID()
	log.WithFields(logrus.Fields{
		"request_id": requestID,
		"path":       req.Path,
	}).Info("Received fetch quote request")

	startTime := time.Now()
	c := colly.NewCollector()
	url := baseURL

	if req.Path != "" {
		if strings.HasPrefix(req.Path, "/") {
			url = baseURL + strings.TrimPrefix(req.Path, "/")
		} else {
			url = req.Path
		}
	}

	log.WithFields(logrus.Fields{
		"request_id": requestID,
		"url":        url,
	}).Debug("Fetching URL")

	c.OnResponse(func(r *colly.Response) {
		log.WithFields(logrus.Fields{
			"request_id": requestID,
			"status":     r.StatusCode,
			"url":        r.Request.URL.String(),
		}).Debug("Received response")
	})

	c.OnHTML(".quote", func(e *colly.HTMLElement) {
		resp.Quote = e.Text
		log.WithFields(logrus.Fields{
			"request_id": requestID,
			"quote":      resp.Quote,
		}).Debug("Found quote")
	})

	c.OnHTML(".links", func(e *colly.HTMLElement) {
		e.ForEach("a", func(_ int, el *colly.HTMLElement) {
			if el.Text == "Excuse suivante" {
				resp.NextLink = el.Attr("href")
				log.WithFields(logrus.Fields{
					"request_id": requestID,
					"next_link":  resp.NextLink,
				}).Debug("Found next link")
			} else if el.Text == "Excuse précédente" {
				resp.PreviousLink = el.Attr("href")
				log.WithFields(logrus.Fields{
					"request_id": requestID,
					"prev_link":  resp.PreviousLink,
				}).Debug("Found previous link")
			}
		})
	})

	c.OnError(func(r *colly.Response, err error) {
		log.WithFields(logrus.Fields{
			"request_id": requestID,
			"url":        r.Request.URL.String(),
			"error":      err,
		}).Error("Error during fetch")
	})

	err := c.Visit(url)
	if err != nil {
		log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err,
		}).Error("Failed to fetch quote")
		resp.Error = err.Error()
		return err
	}

	duration := time.Since(startTime)
	log.WithFields(logrus.Fields{
		"request_id": requestID,
		"duration":   duration.String(),
	}).Info("Request completed successfully")

	return nil
}

func generateRequestID() string {
	return time.Now().Format("20060102-150405") + "-" +
		strings.ReplaceAll(strings.ToLower(uuid.New().String()), "-", "")[:8]
}
