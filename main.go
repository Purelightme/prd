package main

import (
	"embed"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
	tmpl := template.Must(template.New("").ParseFS(f,"views/*"))
	r.SetHTMLTemplate(tmpl)
	r.GET("/", func(c *gin.Context) {
		fmt.Println(tmpl.DefinedTemplates())
		c.HTML(http.StatusOK,"index.html",nil)
	})
	
	r.POST("/upload", func(c *gin.Context) {
		f,err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest,gin.H{
				"error": err.Error(),
			})
			return
		}

		dst := path.Join(*uploadsDir,f.Filename)
		err = c.SaveUploadedFile(f,dst)
		if err != nil {
			c.JSON(http.StatusInternalServerError,gin.H{
				"error": err.Error(),
			})
			return
		}

		//检测文件类型
		contentType := f.Header.Get("Content-Type")
		fmt.Println("文件类型:",contentType)
		url := ""
		switch contentType {
		case "application/zip":
			_,err = tools.Unzip(dst,*wwwDir)
			if err != nil {
				c.JSON(http.StatusInternalServerError,gin.H{
					"error": err.Error(),
				})
				return
			}
			url = strings.Join([]string{"http://"+*host+":"+*port,"/prd/",strings.Split(f.Filename,".")[0],"/index.html"},"")
		default:
			dst,err := os.OpenFile(*wwwDir + "/" + f.Filename,os.O_CREATE|os.O_RDWR,0777)
			if err != nil {
				fmt.Println("文件创建失败",err)
				return
			}
			src,_ := f.Open()
			n,err := io.Copy(dst,src)
			if err != nil {
				fmt.Println("复制失败",err)
				return
			}
			fmt.Println("成功复制",n)
			url = strings.Join([]string{"http://"+*host+":"+*port,"/prd/",f.Filename},"")
		}

		c.JSON(http.StatusOK,gin.H{
			"msg": "upload success",
			"url": url,
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
					Url: strings.Join([]string{"http://"+*host+":"+*port,"/prd/",f.Name()},""),
				})
			}else if !f.IsDir() {
				dirs = append(dirs, Item{
					Name: f.Name(),
					Url: strings.Join([]string{"http://"+*host+":"+*port,"/prd/",f.Name()},""),
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
