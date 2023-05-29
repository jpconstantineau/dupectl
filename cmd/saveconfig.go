/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// saveconfigCmd represents the saveconfig command
var saveconfigCmd = &cobra.Command{
	Use:   "save",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		viper.SetConfigType("yaml")
		viper.SetConfigName(".dupectl.yaml")
		viper.WriteConfigAs(".dupectl.yaml")
		fmt.Println("viper.WriteConfig called - checking files")

		if err := viper.ReadInConfig(); err == nil {
			fmt.Fprintln(os.Stderr, "Using config file: ", viper.ConfigFileUsed())
		} else {
			fmt.Fprintln(os.Stderr, "Error ", error(err))
		}
	},
}

func init() {
	configCmd.AddCommand(saveconfigCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// saveconfigCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// saveconfigCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
