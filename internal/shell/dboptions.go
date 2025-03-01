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
			desc: `Include the trailing semicolon the usql uses to identify the end of a statement.
usql normally includes them, but usqlgen doesn't by default because most Go drivers
don't need or want trailing semicolons`,
			apply: func(input *gen.Input) {
				input.IncludeSemicolon = true
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
	for _, opt := range allOptions {
		err := writeOpt(opt, c.App.Writer)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeOpt(opt *dboption, writer io.Writer) error {
	_, err := fmt.Fprint(writer, opt.name, "\n", "\n")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(writer, opt.desc)
	return err
}
