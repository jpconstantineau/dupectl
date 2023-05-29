/*
Copyright © 2023 Pierre Constantineau jpconstantineau@gmail.com

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/jpconstantineau/dupectl/pkg/auth"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dupectl",
	Short: "Duplicate File Manager",
	Long: `
	Duplicate File Manager        
	 _____                    _____ _______ _      
	|  __ \                  / ____|__   __| |     
	| |  | |_   _ _ __   ___| |       | |  | |     
	| |  | | | | | '_ \ / _ \ |       | |  | |     
	| |__| | |_| | |_) |  __/ |____   | |  | |____ 
	|_____/ \__,_| .__/ \___|\_____|  |_|  |______|
                     | |                               
                     |_|                               			
	
	
Enables searching for Duplicate Files and manage their retention.
Requires configuration file with database connection settings`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.dupectl.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func setDefaults() {
	viper.SetDefault("database.username", "root")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.hostname", "127.0.0.1")
	viper.SetDefault("database.port", "3306")
	viper.SetDefault("database.dbname", "dupedb")

	viper.SetDefault("server.port", "3000")
	viper.SetDefault("server.apikey", auth.GenerateAPISeed())
	viper.SetDefault("server.serverid", auth.GenerateMachineID())

	viper.SetDefault("client.apiport", "3000")
	viper.SetDefault("client.uniqueid", auth.GenerateAPISeed())
	viper.SetDefault("client.clientid", auth.GenerateMachineID())
	viper.SetDefault("client.apikey", "")
	viper.SetDefault("client.apitoken", "")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".dupectl" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".dupectl.yaml")
	}
	setDefaults()

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file: ", viper.ConfigFileUsed())
	}
}
