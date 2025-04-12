package shell

import (
	"fmt"
	"io"
	"strings"

	"github.com/samber/lo"
	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/urfave/cli/v2"
)

type dboption struct {
	name  string
	desc  string
	apply func(input *gen.Input)
}

var (
	allOptions = []*dboption{
		{
			name: "includesemicolon",
			desc: `Include the trailing semicolon that usql uses to identify the end of a statement,
in the SQL string sent to the driver. usql normally includes them, but usqlgen doesn't by
default because most Go drivers don't need or want trailing semicolons.`,
			apply: func(input *gen.Input) {
				input.IncludeSemicolon = true
			},
		},
		{
			name: "keepcgo",
			desc: `Don't replace drivers that require CGO if CGO is not available. Compilation will likely fail.
Useful if generation happens in one environment but compilation in another. See docs for details.`,
			apply: func(input *gen.Input) {
				input.KeepCgo = true
			},
		},
	}
	optionNames = lo.SliceToMap(allOptions, func(o *dboption) (string, *dboption) {
		return o.name, o
	})
)

func fromNames(names []string) ([]*dboption, error) {
	var options []*dboption
	for _, name := range names {
		name = strings.ToLower(name)
		opt, ok := optionNames[name]
		if !ok {
			return nil, fmt.Errorf("unknown option %s", name)
		}
		options = append(options, opt)
	}
	return options, nil
}

func applyOptionsFromNames(names []string, genInput *gen.Input) error {
	activeOpts, err := fromNames(names)
	if err != nil {
		return err
	}
	lo.ForEach(activeOpts, func(item *dboption, _ int) {
		item.apply(genInput)
	})
	return nil
}

func listOptions(c *cli.Context) error {
	_, err := fmt.Fprint(c.App.Writer, "Options available for --dboptions parameter:\n\n")
	if err != nil {
		return err
	}

	for _, opt := range allOptions {
		err = writeOpt(opt, c.App.Writer)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeOpt(opt *dboption, writer io.Writer) error {
	_, err := fmt.Fprint(writer, "- ", opt.name, "\n", "\n")
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(writer, opt.desc, "\n", "\n")
	return err
}
