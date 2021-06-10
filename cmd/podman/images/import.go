package images

import (
	"context"
	"fmt"
	"strings"

	"github.com/containers/common/pkg/completion"
	"github.com/containers/podman/v3/cmd/podman/common"
	"github.com/containers/podman/v3/cmd/podman/parse"
	"github.com/containers/podman/v3/cmd/podman/registry"
	"github.com/containers/podman/v3/pkg/domain/entities"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	importDescription = `Create a container image from the contents of the specified tarball (.tar, .tar.gz, .tgz, .bzip, .tar.xz, .txz).

  Note remote tar balls can be specified, via web address.
  Optionally tag the image. You can specify the instructions using the --change option.`
	importCommand = &cobra.Command{
		Use:               "import [options] PATH [REFERENCE]",
		Short:             "Import a tarball to create a filesystem image",
		Long:              importDescription,
		RunE:              importCon,
		Args:              cobra.RangeArgs(1, 2),
		ValidArgsFunction: common.AutocompleteDefaultOneArg,
		Example: `podman import http://example.com/ctr.tar url-image
  cat ctr.tar | podman -q import --message "importing the ctr.tar tarball" - image-imported
  cat ctr.tar | podman import -`,
	}

	imageImportCommand = &cobra.Command{
		Use:               importCommand.Use,
		Short:             importCommand.Short,
		Long:              importCommand.Long,
		RunE:              importCommand.RunE,
		Args:              importCommand.Args,
		ValidArgsFunction: importCommand.ValidArgsFunction,
		Example: `podman image import http://example.com/ctr.tar url-image
  cat ctr.tar | podman -q image import --message "importing the ctr.tar tarball" - image-imported
  cat ctr.tar | podman image import -`,
	}
)

var (
	importOpts entities.ImageImportOptions
)

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: importCommand,
	})
	importFlags(importCommand)

	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: imageImportCommand,
		Parent:  imageCmd,
	})
	importFlags(imageImportCommand)
}

func importFlags(cmd *cobra.Command) {
	flags := cmd.Flags()

	changeFlagName := "change"
	flags.StringArrayVarP(&importOpts.Changes, changeFlagName, "c", []string{}, "Apply the following possible instructions to the created image (default []): "+strings.Join(common.ChangeCmds, " | "))
	_ = cmd.RegisterFlagCompletionFunc(changeFlagName, common.AutocompleteChangeInstructions)

	messageFlagName := "message"
	flags.StringVarP(&importOpts.Message, messageFlagName, "m", "", "Set commit message for imported image")
	_ = cmd.RegisterFlagCompletionFunc(messageFlagName, completion.AutocompleteNone)

	flags.BoolVarP(&importOpts.Quiet, "quiet", "q", false, "Suppress output")
	flags.StringVar(&importOpts.SignaturePolicy, "signature-policy", "", "Path to a signature-policy file")
	_ = flags.MarkHidden("signature-policy")
}

func importCon(cmd *cobra.Command, args []string) error {
	var (
		source    string
		reference string
	)
	switch len(args) {
	case 0:
		return errors.Errorf("need to give the path to the tarball, or must specify a tarball of '-' for stdin")
	case 1:
		source = args[0]
	case 2:
		source = args[0]
		// TODO when save is merged, we need to process reference
		// like it is done in there or we end up with docker.io prepends
		// instead of the localhost ones
		reference = args[1]
	default:
		return errors.Errorf("too many arguments. Usage TARBALL [REFERENCE]")
	}
	errFileName := parse.ValidateFileName(source)
	errURL := parse.ValidURL(source)
	if errURL == nil {
		importOpts.SourceIsURL = true
	}
	if errFileName != nil && errURL != nil {
		return multierror.Append(errFileName, errURL)
	}

	importOpts.Source = source
	importOpts.Reference = reference

	response, err := registry.ImageEngine().Import(context.Background(), importOpts)
	if err != nil {
		return err
	}
	fmt.Println(response.Id)
	return nil
}
