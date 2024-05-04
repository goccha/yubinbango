package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/goccha/envar"
	"github.com/goccha/fileloaders"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/goccha/problems"
)

func Get(callback, dirPath string) gin.HandlerFunc {
	type Request struct {
		ZipCode  string `uri:"zip" binding:"required,min=7,max=12"`
		Callback string `form:"callback" binding:"omitempty,min=1,max=64"`
	}
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		req := &Request{
			ZipCode: c.Param("zip"),
		}
		if err := c.ShouldBindQuery(req); err != nil {
			problems.New(problems.Path(c.Request), problems.ValidationErrors(err)).BadRequest("").JSON(ctx, c.Writer)
			return
		}
		if req.Callback == "" {
			req.Callback = callback
		}
		zipCode := req.ZipCode // 郵便番号
		ext := ""              // 拡張子
		if index := strings.Index(zipCode, "."); index >= 0 {
			zipCode = zipCode[:index]
			ext = req.ZipCode[index:]
		}
		jsonp := false
		js := false
		if req.Callback != "" {
			jsonp = true
		}
		key := zipCode[:3]
		path := dirPath
		if path == "" {
			path = envar.Get("DATA_DIR_PATH").String("file://data/output/")
		}
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}
		switch ext {
		case ".js":
			path += "js/" + key + ".js"
			js = true
		default:
			path += "json/" + key + ".json"
		}
		bin, err := fileloaders.Load(ctx, path)
		if err != nil {
			problems.New(problems.Path(c.Request)).InternalServerError(err.Error()).JSON(ctx, c.Writer)
			return
		}
		body := string(bin)
		if strings.HasPrefix(body, "$yubin(") {
			body = body[7 : len(body)-2]
			bin = []byte(body)
		}
		res := make(map[string]any)
		if err = json.Unmarshal(bin, &res); err != nil {
			problems.New(problems.Path(c.Request)).InternalServerError("").JSON(ctx, c.Writer)
			return
		}
		if v, ok := res[zipCode]; !ok {
			problems.New(problems.Path(c.Request)).NotFound("").JSON(ctx, c.Writer)
			return
		} else {
			if js {
				v = map[string]any{
					zipCode: v,
				}
			}
			if jsonp {
				bin, err = json.Marshal(v)
				if err != nil {
					problems.New(problems.Path(c.Request)).InternalServerError("").JSON(ctx, c.Writer)
					return
				}
				data := new(bytes.Buffer)
				if req.Callback != "" {
					data.WriteString(req.Callback)
					data.WriteString("(")
				} else {
					data.WriteString("$yubin(")
				}
				data.Write(bin)
				data.WriteString(")")
				c.Data(200, "application/javascript", data.Bytes())
			} else {
				c.JSON(200, v)
			}
		}
	}
}
