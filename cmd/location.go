package cmd

import "github.com/spf13/cobra"

var locationCmd = &cobra.Command{
	Use:     "location",
	Aliases: []string{"l", "loc"},
	Short:   "Convertor IP location from TXT or CSV to .dat.",
	Long:    `Convertor IP location from TXT or CSV to .dat.`,
	Example: `ip2dat loc -i /to/path/ip2location.txt -o /to/path/ip2location.dat`,
	Run:     locationCmdRun,
}

var (
	locationInputFile  string
	locationOutputFile string
)

func init() {
	locationCmd.PersistentFlags().StringVarP(&locationInputFile, "input", "i", "ip2location.txt", "The ip2location input file path")
	locationCmd.PersistentFlags().StringVarP(&locationOutputFile, "output", "o", "ip2location.dat", "The ip2location output file path")
	rootCmd.AddCommand(locationCmd)
}

func locationCmdRun(_ *cobra.Command, _ []string) {

}
