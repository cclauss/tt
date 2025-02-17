package cmd

import (
	"errors"
	"os"

	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/tarantool/tt/cli/util"
)

// handleCmdErr handles an error returned by command implementation.
// If received error is of an ArgError type, usage help is printed.
func handleCmdErr(cmd *cobra.Command, err error) {
	if err != nil {
		var argError *util.ArgError
		if errors.As(err, &argError) {
			log.Error(argError.Error())
			cmd.Usage()
			os.Exit(1)
		}
		log.Fatalf(err.Error())
	}
}
