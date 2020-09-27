/*
// TODO confirm that this is the license that I want to use
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"github.com/madewithlinux/gcodetools"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
)

// minifyCmd represents the minify command
var minifyCmd = &cobra.Command{
	Use:   "minify",
	Short: "minify a gcode file",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		inputFilename := viper.GetString("input")
		outputFilename := viper.GetString("output")
		removeComments := viper.GetBool("removeComments")
		allowUnknownGcode := viper.GetBool("allowUnknownGcode")

		var inputGcodeBytes []byte
		if inputFilename == "-" {
			inputGcodeBytes, err = ioutil.ReadAll(os.Stdin)
			die(err)
		} else {
			inputGcodeBytes, err = ioutil.ReadFile(inputFilename)
			die(err)
		}
		inputGcodeStr := string(inputGcodeBytes)

		cfg := (&gcodetools.GcodeMinifierConfig{
			RemoveComments:    removeComments,
			AllowUnknownGcode: allowUnknownGcode,
		}).Init()
		state := gcodetools.MachineState{}

		outputGcodeStr, _ := cfg.MinifyGcodeStr(state, inputGcodeStr)

		if outputFilename == "-" {
			_, err = os.Stdout.WriteString(outputGcodeStr)
			die(err)
		} else {
			err = ioutil.WriteFile(outputFilename, []byte(outputGcodeStr), 0644)
			die(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(minifyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// minifyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// minifyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	minifyCmd.Flags().StringP("input", "i", "-", "input file to minify (- for stdin)")
	die(minifyCmd.MarkFlagRequired("input"))
	die(viper.BindPFlag("input", minifyCmd.Flags().Lookup("input")))

	minifyCmd.Flags().StringP("output", "o", "-", "file to write minified output to (- for  stdout)")
	die(minifyCmd.MarkFlagRequired("output"))
	die(viper.BindPFlag("output", minifyCmd.Flags().Lookup("output")))

	minifyCmd.Flags().Bool("removeComments", false, "whether to remove comments from minified files")
	die(viper.BindPFlag("removeComments", minifyCmd.Flags().Lookup("removeComments")))

	minifyCmd.Flags().Bool("allowUnknownGcode", false, "continue on gcode that is not understood by this minifier")
	die(viper.BindPFlag("allowUnknownGcode", minifyCmd.Flags().Lookup("allowUnknownGcode")))

}

func die(err error) {
	if err != nil {
		panic(err)
	}
}
