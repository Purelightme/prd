package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"path"
	"prd/tools"
	"strings"
)

func main()  {

	host := flag.String("host","127.0.0.1","Host")
	port := flag.String("port","8080","Port")
	uploadsDir := flag.String("uploadsDir","./uploads","上传文件存储地址")
	wwwDir := flag.String("wwwDir","./www","www服务地址，即解压地址")

	flag.Parse()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Static("/prd","./www")
	r.LoadHTMLFiles("./views/index.html")
	r.GET("/index", func(c *gin.Context) {
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
			"url": path.Join("http://"+*host+":"+*port,"/prd/",strings.Split(f.Filename,".")[0],"/index.html"),
		})
	})
	err := r.Run("0.0.0.0:" + *port)
	if err != nil {
		log.Fatal(err)
	}
}

