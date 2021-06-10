package images

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/containers/common/pkg/completion"
	"github.com/containers/podman/v3/cmd/podman/registry"
	"github.com/containers/podman/v3/cmd/podman/validate"
	"github.com/containers/podman/v3/pkg/domain/entities"
	"github.com/containers/podman/v3/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	loadDescription = "Loads an image from a locally stored archive (tar file) into container storage."
	loadCommand     = &cobra.Command{
		Use:               "load [options]",
		Short:             "Load image(s) from a tar archive",
		Long:              loadDescription,
		RunE:              load,
		Args:              validate.NoArgs,
		ValidArgsFunction: completion.AutocompleteNone,
	}

	imageLoadCommand = &cobra.Command{
		Args:              loadCommand.Args,
		Use:               loadCommand.Use,
		Short:             loadCommand.Short,
		Long:              loadCommand.Long,
		ValidArgsFunction: loadCommand.ValidArgsFunction,
		RunE:              loadCommand.RunE,
	}
)

var (
	loadOpts entities.ImageLoadOptions
)

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: loadCommand,
	})
	loadFlags(loadCommand)
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: imageLoadCommand,
		Parent:  imageCmd,
	})
	loadFlags(imageLoadCommand)
}

func loadFlags(cmd *cobra.Command) {
	flags := cmd.Flags()

	inputFlagName := "input"
	flags.StringVarP(&loadOpts.Input, inputFlagName, "i", "", "Read from specified archive file (default: stdin)")
	_ = cmd.RegisterFlagCompletionFunc(inputFlagName, completion.AutocompleteDefault)

	flags.BoolVarP(&loadOpts.Quiet, "quiet", "q", false, "Suppress the output")
	flags.StringVar(&loadOpts.SignaturePolicy, "signature-policy", "", "Pathname of signature policy file")
	_ = flags.MarkHidden("signature-policy")
}

func load(cmd *cobra.Command, args []string) error {
	if len(loadOpts.Input) > 0 {
		if _, err := os.Stat(loadOpts.Input); err != nil {
			return err
		}
	} else {
		if terminal.IsTerminal(int(os.Stdin.Fd())) {
			return errors.Errorf("cannot read from terminal. Use command-line redirection or the --input flag.")
		}
		outFile, err := ioutil.TempFile(util.Tmpdir(), "podman")
		if err != nil {
			return errors.Errorf("error creating file %v", err)
		}
		defer os.Remove(outFile.Name())
		defer outFile.Close()

		_, err = io.Copy(outFile, os.Stdin)
		if err != nil {
			return errors.Errorf("error copying file %v", err)
		}
		loadOpts.Input = outFile.Name()
	}
	response, err := registry.ImageEngine().Load(context.Background(), loadOpts)
	if err != nil {
		return err
	}
	fmt.Println("Loaded image(s): " + strings.Join(response.Names, ","))
	return nil
}
