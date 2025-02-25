package cmd

import "github.com/spf13/cobra"

var asnCmd = &cobra.Command{
	Use:     "asn",
	Aliases: []string{},
	Short:   "Convertor IP asn from TXT or CSV to .dat.",
	Long:    `Convertor IP asn from TXT or CSV to .dat.`,
	Example: `ip2dat asn -i /to/path/ip2asn.txt -o /to/path/ip2asn.dat`,
	Run:     asnCmdRun,
}

var (
	asnInputFile  string
	asnOutputFile string
)

func init() {
	asnCmd.PersistentFlags().StringVarP(&asnInputFile, "input", "i", "ip2asn.txt", "The ip2asn input file path")
	asnCmd.PersistentFlags().StringVarP(&asnOutputFile, "output", "o", "ip2asn.dat", "The ip2asn output file path")
	rootCmd.AddCommand(asnCmd)
}

func asnCmdRun(_ *cobra.Command, _ []string) {

}
