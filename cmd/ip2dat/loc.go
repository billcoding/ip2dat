package main

import (
	"fmt"

	"github.com/billcoding/ip2dat/ip2loc"
	"github.com/billcoding/ip2dat/iplocsearch"
	"github.com/spf13/cobra"
)

var locationCmd = &cobra.Command{
	Use:     "location",
	Aliases: []string{"l", "loc"},
	Short:   "Convertor IP location from TXT or CSV to .dat.",
	Long:    `Convertor IP location from TXT or CSV to .dat.`,
	Example: `ip2dat loc -i /to/path/ip2location.txt -o /to/path/ip2location.dat`,
	Run: func(_ *cobra.Command, _ []string) {
		ip2loc.Convert(locationInputFile, locationOutputFile)
		if locationTest && locationTestIp != "" {
			fmt.Println(locationTestIp + " location: " + iplocsearch.Search(locationOutputFile, locationTestIp))
		}
	}}

var (
	locationInputFile  string
	locationOutputFile string
	locationTest       bool
	locationTestIp     string
)

func init() {
	locationCmd.PersistentFlags().StringVarP(&locationInputFile, "input", "i", "ip2loc.txt", "The ip2location input file path")
	locationCmd.PersistentFlags().StringVarP(&locationOutputFile, "output", "o", "ip2loc.dat", "The ip2location output file path")
	locationCmd.PersistentFlags().BoolVarP(&locationTest, "test", "t", false, "Test after converted")
	locationCmd.PersistentFlags().StringVar(&locationTestIp, "test-ip", "1.1.1.1", "Test ip address")
	rootCmd.AddCommand(locationCmd)
}
