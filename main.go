package main

import (
	"TRAFforward/src"
	"TRAFforward/src/common"
	"TRAFforward/src/database"
	"TRAFforward/src/models"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	//go:embed www/js
	JSEmbed embed.FS

	//go:embed www/css
	CSSEmbed embed.FS

	//go:embed www/*
	templatesEmbed embed.FS
)

func InitForward() {
	db := database.GetDB()
	list := models.ForwardGetPortList(db)
	for _, v := range list {
		ok, port := common.ValidatePort(v.BindPort)
		if !ok {
			fmt.Printf("配置异常或端口[%s]已被占用", port)
			continue
			// return
		}
		go common.RunTransferred(0, 1, port, v.Destination, func(use_total uint64) {
			models.ForwardUpdateUse(db, v.ID, use_total)
		})
	}
}

func main() {
	r := gin.Default()
	templ := template.Must(template.New("").ParseFS(templatesEmbed, "www/*.html"))
	r.SetHTMLTemplate(templ)
	jsFS, _ := fs.Sub(JSEmbed, "www/js")
	r.StaticFS("/js", http.FS(jsFS))
	cssFS, _ := fs.Sub(CSSEmbed, "www/css")
	r.StaticFS("/css", http.FS(cssFS))

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})
	cfg := database.InitConfig()
	db := database.InitDB(cfg)
	db.AutoMigrate(
		&models.User{},
		&models.Forward{},
	)
	models.UserCreateAdmin(db)
	src.RouterRegister(r)
	InitForward()
	r.Run(cfg.Addr)
}
