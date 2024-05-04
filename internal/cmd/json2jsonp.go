package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccha/yubinbango/pkg/entities"

	"github.com/goccha/logging/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(Json2Jsonp)
}

var Json2Jsonp = NewJson2Jsonp()

func NewJson2Jsonp() *cobra.Command {
	type Options struct {
		Path   string
		Output string
	}
	options := &Options{}
	cmd := &cobra.Command{
		Use:     "json2jsonp",
		Aliases: []string{"j2j"},
		Short:   "Convert data from json to jsonp",
		Long:    "Convert data from json to jsonp",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			files, err := os.ReadDir(options.Path)
			if err != nil {
				return err
			}
			for _, file := range files {
				if file.IsDir() {
					continue
				}
				if strings.HasSuffix(file.Name(), ".json") {
					if err = convert(ctx, options.Path, file.Name(), options.Output); err != nil {
						return err
					}
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&options.Path, "path", "p", "./data/output/json", "Path to load json from")
	cmd.Flags().StringVarP(&options.Output, "output", "o", "./data/output/js", "Output path")
	return cmd
}

func convert(ctx context.Context, path, fileName, output string) error {
	if !strings.HasSuffix(path, "/json") && !strings.HasSuffix(path, "/json/") {
		path = filepath.Join(path, "json")
	}
	f, err := entities.OpenFile(ctx, path, fileName)
	if err != nil {
		log.Fatal(ctx).Msgf("read: %+v", err)
		return err
	}
	if !strings.HasSuffix(output, "/js") && !strings.HasSuffix(output, "/js/") {
		output = filepath.Join(output, "js")
	}
	if err := os.MkdirAll(output, os.ModePerm); err != nil {
		log.Fatal(ctx).Msgf("mkdir: %+v", err)
		return err
	}
	format := &entities.JsFormat{}
	v, err := format.Format(f)
	if err != nil {
		log.Fatal(ctx).Msgf("format: %+v")
		return err
	}
	fileName = strings.Replace(fileName, ".json", ".js", 1)
	file, err := os.Create(filepath.Join(output, fileName))
	if err != nil {
		log.Fatal(ctx).Err(err).Send()
		return err
	}
	defer func() {
		_ = file.Close()
	}()
	if _, err = file.Write([]byte(v)); err != nil {
		log.Fatal(ctx).Err(err).Send()
	}
	return nil
}
