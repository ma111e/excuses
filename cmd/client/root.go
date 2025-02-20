package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ma111e/excuses/internal/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var (
	cfgFile    string
	serverAddr string
)

var rootCmd = &cobra.Command{
	Use:   "client",
	Short: "A TUI for cyber security excuses",
	Long:  `A terminal user interface that displays excuses from https://cyber.excusesecu.fr/`,
	Run: func(_ *cobra.Command, _ []string) {
		p := tea.NewProgram(tui.InitialModel(serverAddr))
		_, err := p.Run()
		cobra.CheckErr(err)
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./excuses-client.yml)")
	rootCmd.PersistentFlags().StringVar(&serverAddr, "server", "localhost:1234", "RPC server address")

	_ = viper.BindPFlag("server", rootCmd.PersistentFlags().Lookup("server"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.SetConfigName("excuses-client")
		viper.SetConfigType("yml")
		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	if viper.IsSet("server") {
		serverAddr = viper.GetString("server")
	}
}

func Execute() error {
	return rootCmd.Execute()
}
