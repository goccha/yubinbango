package cmd

import (
	"bufio"
	"context"
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/goccha/yubinbango/pkg/entities"
	"github.com/goccha/yubinbango/pkg/parsers"

	"github.com/goccha/logging/log"
	"github.com/spf13/cobra"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func init() {
	rootCmd.AddCommand(Csv2Json)
}

var Csv2Json = NewCsv2Json()

func NewCsv2Json() *cobra.Command {
	type Options struct {
		Paths  string
		Output string
		Renew  bool
	}
	options := &Options{}
	cmd := &cobra.Command{
		Use:     "csv2json",
		Aliases: []string{"c2j"},
		Short:   "Load data from csv and output as json",
		Long:    "Load data from csv and output as json",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			parser := parsers.NewParser()
			filePaths, err := parsePath(options.Paths)
			if err != nil {
				return err
			}
			var m map[string]*entities.File
			for _, path := range filePaths {
				m, err = load(ctx, path, parser, m)
				if err != nil {
					return err
				}
			}
			if err = writeJson(ctx, m, options.Output, options.Renew); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&options.Paths, "path", "p", "./data/*.csv,./data/*.CSV", "Path to load data from")
	cmd.Flags().StringVarP(&options.Output, "output", "o", "./data/output/json", "Output path")
	cmd.Flags().BoolVarP(&options.Renew, "renew", "r", false, "Renew output directory")
	return cmd
}

func parsePath(filePath string) ([]string, error) {
	var files []string
	pathList := strings.Split(filePath, ",")
	for _, path := range pathList {
		matches, err := filepath.Glob(path)
		if err != nil {
			return nil, err
		}
		files = append(files, matches...)
	}
	return files, nil
}

func load(ctx context.Context, path string, parser parsers.Parser, m map[string]*entities.File) (map[string]*entities.File, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(ctx).Msgf("open: %+v", err)
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()
	r, err := reader(ctx, file)
	if err != nil {
		return nil, err
	}
	if m == nil {
		m = make(map[string]*entities.File)
	}
	for {
		if row, err := r.Read(); err != nil {
			if err == io.EOF {
				break
			}
		} else {
			entity := parser.Parse(ctx, row)
			key := entity.ZipCode[:3]
			if v, ok := m[key]; ok {
				v.Add(ctx, &entity)
			} else {
				f := &entities.File{Key: key, Ext: "json"}
				m[key] = f
				f.Add(ctx, &entity)
			}
		}
	}
	return m, nil
}

func reader(ctx context.Context, fp *os.File) (*csv.Reader, error) {
	scanner := bufio.NewScanner(fp)
	if scanner.Scan() {
		line := scanner.Text()
		if !utf8.ValidString(line) {
			if _, err := fp.Seek(0, 0); err != nil {
				return nil, err
			}
			sjis := transform.NewReader(fp, japanese.ShiftJIS.NewDecoder())
			return csv.NewReader(sjis), nil
		}
	}
	if _, err := fp.Seek(0, 0); err != nil {
		return nil, err
	}
	return csv.NewReader(fp), nil
}

func writeJson(ctx context.Context, m map[string]*entities.File, output string, renew bool) error {
	if !strings.HasSuffix(output, "/json") && !strings.HasSuffix(output, "/json/") {
		output = filepath.Join(output, "json")
	}
	if err := os.MkdirAll(output, 0755); err != nil {
		return err
	}
	for _, v := range m {
		if err := v.Write(ctx, output, renew); err != nil {
			return err
		}
	}
	return nil
}
