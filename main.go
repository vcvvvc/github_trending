package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"

	"github.com/andygrunwald/go-trending"
	"golang.org/x/net/proxy"
)

type Item struct {
	ID          int
	URL         string
	Name        string
	Language    string
	Stars       int
	Description string
}

type LanguageGroup struct {
	GroupName string
	Repos     []Item
}

func open(url string) {
	err := exec.Command("open", url).Start()
	if err != nil {
		fmt.Println("Failed to open the URL:", err)
	}
}

// 将HTML内容保存到指定的文件
func saveHTMLToFile(filename string, content string) error {
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		log.Fatalf("failed to create file: %v", err)
		return err
	}
	return nil
}

// 渲染模板并保存到文件
func renderTemplateToFile(templateFile string, data interface{}, outputFilename string) error {
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		return fmt.Errorf("Error parsing template: %v", err)
	}

	var htmlBuffer bytes.Buffer
	err = tmpl.Execute(&htmlBuffer, data)
	if err != nil {
		return fmt.Errorf("Error rendering template: %v", err)
	}

	// 确保目录存在
	err = os.MkdirAll("daily_trending", os.ModePerm)
	if err != nil {
		return fmt.Errorf("Error creating directory: %v", err)
	}
	
	err = saveHTMLToFile(outputFilename, htmlBuffer.String())
	if err != nil {
		return fmt.Errorf("Error saving HTML to file: %v", err)
	}

	return nil
}

func main() {
	todayStr := time.Now().Format("2006-01-02")
	dailyFilename := fmt.Sprintf("daily_trending/%s.html", todayStr)
	
	var client *http.Client
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		// In GitHub Actions, use a default HTTP client.
		client = &http.Client{}
	} else {
		// For local development, use a SOCKS5 proxy.
		socks5URL, err := url.Parse("socks5://127.0.0.1:8800")
		if err != nil {
			log.Fatalf("Failed to parse SOCKS5 URL: %v", err)
		}

		dialer, err := proxy.FromURL(socks5URL, proxy.Direct)
		if err != nil {
			log.Println("Can't connect to the proxy, trying direct connection")
			client = &http.Client{} // Fallback to direct connection
		} else {
			httpTransport := &http.Transport{
				Dial: dialer.Dial,
			}
			client = &http.Client{Transport: httpTransport}
		}
	}

	trend := trending.NewTrendingWithClient(client)
	var groupedItems []LanguageGroup

	lists := []string{
		"", "C++", "Go", "Python", "Solidity", "Rust", 
	}

	idCounter := 1
	for _, langName := range lists {
		fmt.Printf("Fetching trending projects for: %s\n", langName)

		projects, err := trend.GetProjects(trending.TimeToday, langName)
		if err != nil {
			log.Printf("!!!!!! FAILED to get projects for language '%s': %v !!!!!!", langName, err)
			continue
		}
		if len(projects) == 0 {
			log.Printf("Warning: Found 0 projects for language '%s'.", langName)
			continue
		}

		var groupRepos []Item
		for _, project := range projects {
			repo := Item{
				ID:          idCounter,
				URL:         project.URL.String(),
				Name:        project.Name,
				Language:    project.Language,
				Stars:       project.Stars,
				Description: project.Description,
			}
			groupRepos = append(groupRepos, repo)
			idCounter++
		}

		if len(groupRepos) > 0 {
			groupName := langName
			if groupName == "" {
				groupName = "Overall"
			}
			currentGroup := LanguageGroup{
				GroupName: groupName,
				Repos:     groupRepos,
			}
			groupedItems = append(groupedItems, currentGroup)
		}
		time.Sleep(1 * time.Second)
	}

	if len(groupedItems) == 0 {
		log.Fatalln("CRITICAL: Failed to fetch any projects from GitHub. Aborting.")
	}

	jsonData, err := json.Marshal(groupedItems)
	if err != nil {
		log.Fatalf("Failed to marshal data to JSON: %v", err)
	}

	templateFile := "templates/index.tmpl"
	data := map[string]interface{}{
		"Today":      todayStr,
		"Year":       time.Now().Year(),
		"GroupsJSON": template.JS(jsonData),
	}

	// Render to the root index.html first
	err = renderTemplateToFile(templateFile, data, "index.html")
	if err != nil {
		fmt.Printf("Error rendering root index.html: %v\n", err)
	} else {
		fmt.Println("Root index.html updated successfully.")
	}

	// Also save a copy to the daily archive
	err = renderTemplateToFile(templateFile, data, dailyFilename)
	if err != nil {
		fmt.Printf("Error rendering daily archive file: %v\n", err)
	} else {
		fmt.Println("Daily archive file saved successfully:", dailyFilename)
	}

	// 生成daily_trending/index.html文件列表
	err = GenerateDailyIndex()
	if err != nil {
		fmt.Printf("Error generating daily_trending/index.html: %v\n", err)
	} else {
		fmt.Println("Daily_trending/index.html updated successfully.")
	}

	if os.Getenv("GITHUB_ACTIONS") != "true" {
		open("index.html")
	}
}
