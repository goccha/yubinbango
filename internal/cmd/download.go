package cmd

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	kenAllUrl   = "https://www.post.japanpost.jp/zipcode/dl/utf/zip/utf_ken_all.zip"
	jigyosyoUrl = "https://www.post.japanpost.jp/zipcode/dl/jigyosyo/zip/jigyosyo.zip"
)

func init() {
	rootCmd.AddCommand(Download)
}

var Download = NewDownload()

func NewDownload() *cobra.Command {
	type Options struct {
		KenAll    string
		Jigyosyo  string
		OutputDir string
	}
	options := &Options{}
	cmd := &cobra.Command{
		Use:     "download",
		Aliases: []string{"dl"},
		Short:   "Download file from url",
		Long:    "Download file from url",
		RunE: func(cmd *cobra.Command, args []string) error {
			kenAll := options.KenAll
			if kenAll == "" {
				kenAll = kenAllUrl
			}
			if err := downloadFile(kenAll, options.OutputDir); err != nil {
				return err
			}
			jigyosyo := options.Jigyosyo
			if jigyosyo == "" {
				jigyosyo = jigyosyoUrl
			}
			if err := downloadFile(jigyosyo, options.OutputDir); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&options.KenAll, "ken-all", "k", "", "Download url for ken_all.zip")
	cmd.Flags().StringVarP(&options.Jigyosyo, "jigyosyo", "j", "", "Download url for jigyosyo.zip")
	cmd.Flags().StringVarP(&options.OutputDir, "output-dir", "o", "data", "Output directory")
	return cmd
}

func downloadFile(url string, filepath string) error {
	fmt.Printf("download %s\n", url)
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		_ = res.Body.Close()
	}()
	buf := new(bytes.Buffer)
	if _, err = io.Copy(buf, res.Body); err != nil {
		return err
	}
	if !strings.HasSuffix(filepath, "/") {
		filepath += "/"
	}
	r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		return err
	}
	for _, f := range r.File {
		if err = extract(f, filepath+f.Name); err != nil {
			return err
		}
	}
	return nil
}

func extract(f *zip.File, filepath string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer func() {
		_ = rc.Close()
	}()
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()
	if _, err = io.Copy(out, rc); err != nil {
		return err
	}
	return nil
}
