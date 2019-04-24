// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "gosaas",
	Long: `
---------------------------------------------------
   .aMMMMP .aMMMb  .dMMMb  .aMMMb  .aMMMb  .dMMMb
  dMP"    dMP"dMP dMP" VP dMP"dMP dMP"dMP dMP" VP
 dMP MMP"dMP dMP  VMMMb  dMMMMMP dMMMMMP  VMMMb  
dMP.dMP dMP.aMP dP .dMP dMP dMP dMP dMP dP .dMP  
VMMMP"  VMMMP"  VMMMP" dMP dMP dMP dMP  VMMMP"   
---------------------------------------------------
`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gosaas.yaml)")
	for _, c := range rootCmd.Commands() {
		_ = viper.BindPFlags(c.PersistentFlags())
		_ = viper.BindPFlags(c.Flags())
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".gosaas" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".gosaas")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
