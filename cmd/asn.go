package cmd

import (
	"fmt"

	"github.com/billcoding/ip2dat/cmd/internal/ip2asn"
	"github.com/billcoding/ip2dat/ipasnsearch"
	"github.com/spf13/cobra"
)

var asnCmd = &cobra.Command{
	Use:     "asn",
	Aliases: []string{},
	Short:   "Convertor IP asn from TXT or CSV to .dat.",
	Long:    `Convertor IP asn from TXT or CSV to .dat.`,
	Example: `ip2dat asn -i /to/path/ip2asn.txt -o /to/path/ip2asn.dat`,
	Run: func(_ *cobra.Command, _ []string) {
		ip2asn.Convert(asnInputFile, asnOutputFile)
		if asnTest && asnTestIp != "" {
			fmt.Println(asnTestIp + " asn: " + ipasnsearch.Search(asnOutputFile, asnTestIp))
		}
	},
}

var (
	asnInputFile  string
	asnOutputFile string
	asnTest       bool
	asnTestIp     string
)

func init() {
	asnCmd.PersistentFlags().StringVarP(&asnInputFile, "input", "i", "ip2asn.txt", "The ip2asn input file path")
	asnCmd.PersistentFlags().StringVarP(&asnOutputFile, "output", "o", "ip2asn.dat", "The ip2asn output file path")
	asnCmd.PersistentFlags().BoolVarP(&asnTest, "test", "t", false, "Test after converted")
	asnCmd.PersistentFlags().StringVar(&asnTestIp, "test-ip", "1.1.1.1", "Test ip address")
	rootCmd.AddCommand(asnCmd)
}
