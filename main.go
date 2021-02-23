package main

import (
	"embed"
	"flag"
	"github.com/gin-gonic/gin"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"prd/tools"
	"strings"
)

//go:embed views/*
var f embed.FS

func main()  {

	host := flag.String("host","127.0.0.1","Host")
	port := flag.String("port","8080","Port")
	uploadsDir := flag.String("uploadsDir","./uploads","上传文件存储地址")
	wwwDir := flag.String("wwwDir","./www","www服务地址，即解压地址")

	flag.Parse()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Static("/prd","./www")
	tmpl := template.Must(template.New("").ParseFS(f,"views/*.html"))
	r.SetHTMLTemplate(tmpl)
	
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK,"index.html",nil)
	})
	
	r.POST("/upload", func(c *gin.Context) {
		f,err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest,gin.H{
				"error": err.Error(),
			})
		}

		dst := path.Join(*uploadsDir,f.Filename)
		err = c.SaveUploadedFile(f,dst)
		if err != nil {
			c.JSON(http.StatusInternalServerError,gin.H{
				"error": err.Error(),
			})
		}
		_,err = tools.Unzip(dst,*wwwDir)
		if err != nil {
			c.JSON(http.StatusInternalServerError,gin.H{
				"error": err.Error(),
			})
		}

		c.JSON(http.StatusOK,gin.H{
			"msg": "upload success",
			"url": strings.Join([]string{"http://"+*host+":"+*port,"/prd/",strings.Split(f.Filename,".")[0],"/index.html"},""),
		})
	})

	r.GET("/dirs", func(c *gin.Context) {
		dir,err := ioutil.ReadDir(*wwwDir)
		if err != nil {
			c.JSON(http.StatusInternalServerError,gin.H{
				"error": err.Error(),
			})
		}

		type Item struct {
			Name string `json:"name"`
			Url string `json:"url"`
		}

		var dirs []Item
		for _,f := range dir {
			if f.IsDir() && len(f.Name()) > 0 && f.Name() != "__MACOSX" {
				dirs = append(dirs, Item{
					Name: f.Name(),
					Url: strings.Join([]string{"http://"+*host+":"+*port,"/prd/",f.Name(),"/index.html"},""),
				})
			}
		}

		c.JSON(http.StatusOK,gin.H{
			"dirs": dirs,
		})
	})

	err := r.Run("0.0.0.0:" + *port)
	if err != nil {
		log.Fatal(err)
	}
}