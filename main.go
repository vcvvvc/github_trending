package main

import (
	"bytes"
	"fmt"
	"github.com/andygrunwald/go-trending"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/proxy"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Item struct {
	Id          int
	Url         string
	Name        string
	Languages   string
	Stars       int
	Description string
}

// 将HTML内容保存到指定的文件
func saveHTMLToFile(filename string, content string) error {
	//file, err := os.Create(filename) // 创建文件
	//if err != nil {
	//	return err
	//}
	//defer file.Close()
	//
	//_, err = file.WriteString(content) // 写入内容
	//if err != nil {
	//	return err
	//}
	err := os.MkdirAll("daily_trending", os.ModePerm)
	if err != nil {
		log.Fatalf("failed to create directory: %v", err)
		return err
	}

	err = os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		log.Fatalf("failed to create file: %v", err)
		return err
	}
	return nil
}

// 渲染模板并保存到文件
func renderTemplateToFile(templateFile string, data interface{}, outputFilename string) error {
	// 解析模板文件
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		return fmt.Errorf("Error parsing template: %v", err)
	}

	// 创建一个字节缓冲区来捕获模板渲染的输出
	var htmlBuffer bytes.Buffer
	err = tmpl.Execute(&htmlBuffer, data)
	if err != nil {
		return fmt.Errorf("Error rendering template: %v", err)
	}

	// 将HTML内容保存到指定文件
	err = saveHTMLToFile(outputFilename, htmlBuffer.String())
	if err != nil {
		return fmt.Errorf("Error saving HTML to file: %v", err)
	}

	return nil
}

func startweb(items []Item, outputFilename string) {
	//	// 初始化Gin引擎
	r := gin.Default()
	r.Static("/daily_trending", "./daily_trending")
	//
	//	// 加载HTML模板文件
	//	r.LoadHTMLGlob("templates/*")
	//	// 定义一个GET路由，当访问"/"时，渲染HTML页面
	//	r.GET("/", func(c *gin.Context) {
	//		// 使用HTML模板渲染数据
	//		c.HTML(http.StatusOK, "index.tmpl", gin.H{
	//			"Items": items,
	//		})
	//	})
	templateFile := "templates/index.tmpl"

	// 渲染模板并保存为文件
	err := renderTemplateToFile(templateFile, gin.H{"Items": items}, outputFilename)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("HTML file saved successfully:", outputFilename)
	}

	r.GET("/", func(c *gin.Context) {
		// 读取HTML文件
		data, err := os.ReadFile(outputFilename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 设置响应头
		c.Header("Content-Type", "text/html; charset=utf-8")

		// 返回HTML内容
		c.String(http.StatusOK, string(data))

	})

	fmt.Println("http://127.0.0.1:20111")
	r.Run(":20111")
}

func main() {
	todayStr := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("daily_trending/%s.html", todayStr)
	//fmt.Println(filename)

	// 设置 SOCKS5 代理地址
	socks5URL, _ := url.Parse("socks5://127.0.0.1:1099")

	// 创建代理拨号器
	dialer, err := proxy.FromURL(socks5URL, proxy.Direct)
	if err != nil {
		// 处理错误
	}

	// 设置 http.Transport 使用代理拨号器
	httpTransport := &http.Transport{}
	httpTransport.Dial = dialer.Dial

	// 创建 http.Client 使用定制的 Transport
	client := &http.Client{Transport: httpTransport}
	trend := trending.NewTrendingWithClient(client)

	var items []Item

	// Show projects of today
	lists := []string{
		"",
		"C++",
		"Go",
		"Python",
		"Solidity",
		"Rust",
	}
	for _, list := range lists {
		fmt.Printf("\n\n\n get %s language star list ", list)

		projects, err := trend.GetProjects(trending.TimeToday, list)
		if err != nil {
			panic(err)
		}

		for index, project := range projects {
			i := index + 1
			if len(project.Language) > 0 {
				// 				fmt.Printf("%d: %s\n %s (written in %s with %d 🌟 \n Desc:%s )\n", i, project.URL, project.Name, project.Language, project.Stars, project.Description)
				repo := Item{
					Id:          i,
					Url:         project.URL.String(),
					Name:        project.Name,
					Languages:   project.Language,
					Stars:       project.Stars,
					Description: project.Description,
				}

				items = append(items, repo)

			} else {
				fmt.Printf("%d: %s (with %d ★ )\n", i, project.Name, project.Stars)
			}
		}

		time.Sleep(5)
	}

	startweb(items, filename)
}
