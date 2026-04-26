package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/andygrunwald/go-trending"
	"golang.org/x/net/proxy"
)

type Item struct {
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

type FetchLanguageGroup struct {
	DisplayName  string
	QueryTargets []string
}

type HomePageData struct {
	Today       string
	GroupCount  int
	RepoCount   int
	GroupsJSON  template.JS
	AssetPrefix string
}

type LLMSRepo struct {
	URL         string `json:"url"`
	Name        string `json:"name"`
	Language    string `json:"language"`
	Stars       int    `json:"stars"`
	Description string `json:"description"`
}

type LLMSGroup struct {
	Name      string     `json:"name"`
	RepoCount int        `json:"repo_count"`
	Repos     []LLMSRepo `json:"repos"`
}

type LLMSDocument struct {
	Title      string      `json:"title"`
	Date       string      `json:"date"`
	Source     string      `json:"source"`
	History    string      `json:"history"`
	GroupCount int         `json:"group_count"`
	RepoCount  int         `json:"repo_count"`
	Groups     []LLMSGroup `json:"groups"`
}

func countUniqueReposForHome(groups []LanguageGroup) int {
	seen := make(map[string]struct{})
	for _, group := range groups {
		for _, repo := range group.Repos {
			key := strings.TrimSpace(repo.URL)
			if key == "" {
				key = fmt.Sprintf("%s-%d", repo.Name, repo.Stars)
			}
			seen[key] = struct{}{}
		}
	}
	return len(seen)
}

func open(url string) {
	err := exec.Command("open", url).Start()
	if err != nil {
		fmt.Println("Failed to open the URL:", err)
	}
}

func saveHTMLToFile(filename string, content string) error {
	return os.WriteFile(filename, []byte(content), 0o644)
}

func renderTemplateToFile(templateFile string, data any, outputFilename string) error {
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		return fmt.Errorf("parse template %s: %w", templateFile, err)
	}

	outputDir := filepath.Dir(outputFilename)
	if outputDir != "." {
		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			return fmt.Errorf("create output dir %s: %w", outputDir, err)
		}
	}

	file, err := os.Create(outputFilename)
	if err != nil {
		return fmt.Errorf("create output file %s: %w", outputFilename, err)
	}
	defer file.Close()

	err = tmpl.Execute(file, data)
	if err != nil {
		return fmt.Errorf("execute template %s: %w", templateFile, err)
	}

	return nil
}

func generateLLMSJSON(today string, groups []LanguageGroup) error {
	doc := LLMSDocument{
		Title:      "每日仓库更新",
		Date:       today,
		Source:     "https://0120012.xyz/github_trending/index.html",
		History:    "https://0120012.xyz/github_trending/daily_trending/history.html",
		GroupCount: len(groups),
		RepoCount:  countUniqueReposForHome(groups),
		Groups:     make([]LLMSGroup, 0, len(groups)),
	}

	for _, group := range groups {
		jsonGroup := LLMSGroup{
			Name:      group.GroupName,
			RepoCount: len(group.Repos),
			Repos:     make([]LLMSRepo, 0, len(group.Repos)),
		}

		for _, repo := range group.Repos {
			jsonGroup.Repos = append(jsonGroup.Repos, LLMSRepo{
				URL:         repo.URL,
				Name:        repo.Name,
				Language:    repo.Language,
				Stars:       repo.Stars,
				Description: strings.Join(strings.Fields(repo.Description), " "),
			})
		}

		doc.Groups = append(doc.Groups, jsonGroup)
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal llms.json: %w", err)
	}

	data = append(data, '\n')
	return os.WriteFile("llms.json", data, 0o644)
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
		socks5URL, err := url.Parse("socks5://127.0.0.1:12828")
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

	fetchGroups := []FetchLanguageGroup{
		{DisplayName: "Trending", QueryTargets: []string{""}},
		{DisplayName: "C++", QueryTargets: []string{"C++"}},
		{DisplayName: "Python", QueryTargets: []string{"Python"}},
		{DisplayName: "Go", QueryTargets: []string{"Go"}},
		{DisplayName: "Rust", QueryTargets: []string{"Rust"}},
		{DisplayName: "Java", QueryTargets: []string{"Java"}},
		{DisplayName: "JavaScript/TypeScript", QueryTargets: []string{"JavaScript", "TypeScript"}},
	}

	for _, fetchGroup := range fetchGroups {
		fmt.Printf("Fetching trending projects for group: %s\n", fetchGroup.DisplayName)
		repoByURL := make(map[string]Item)

		for _, queryLang := range fetchGroup.QueryTargets {
			projects, err := trend.GetProjects(trending.TimeToday, queryLang)
			if err != nil {
				log.Printf("!!!!!! FAILED to get projects for language '%s': %v !!!!!!", queryLang, err)
				time.Sleep(1 * time.Second)
				continue
			}
			if len(projects) == 0 {
				log.Printf("Warning: Found 0 projects for language '%s'.", queryLang)
				time.Sleep(1 * time.Second)
				continue
			}

			for _, project := range projects {
				repoURL := project.URL.String()
				if _, exists := repoByURL[repoURL]; exists {
					continue
				}
				repo := Item{
					URL:         repoURL,
					Name:        project.Name,
					Language:    project.Language,
					Stars:       project.Stars,
					Description: project.Description,
				}
				repoByURL[repoURL] = repo
			}
			time.Sleep(1 * time.Second)
		}

		if len(repoByURL) == 0 {
			continue
		}

		groupRepos := make([]Item, 0, len(repoByURL))
		for _, repo := range repoByURL {
			groupRepos = append(groupRepos, repo)
		}
		sort.Slice(groupRepos, func(i, j int) bool {
			return groupRepos[i].Stars > groupRepos[j].Stars
		})

		groupedItems = append(groupedItems, LanguageGroup{
			GroupName: fetchGroup.DisplayName,
			Repos:     groupRepos,
		})
	}

	if len(groupedItems) == 0 {
		log.Fatalln("CRITICAL: Failed to fetch any projects from GitHub. Aborting.")
	}

	jsonData, err := json.Marshal(groupedItems)
	if err != nil {
		log.Fatalf("Failed to marshal data to JSON: %v", err)
	}

	homeData := HomePageData{
		Today:       todayStr,
		GroupCount:  len(groupedItems),
		RepoCount:   countUniqueReposForHome(groupedItems),
		GroupsJSON:  template.JS(jsonData),
		AssetPrefix: "",
	}
	dailyHomeData := homeData
	dailyHomeData.AssetPrefix = "../"

	err = renderTemplateToFile("templates/index.tmpl", homeData, "index.html")
	if err != nil {
		fmt.Printf("Error rendering root index.html: %v\n", err)
	} else {
		fmt.Println("Root index.html updated successfully.")
	}

	err = renderTemplateToFile("templates/index.tmpl", dailyHomeData, dailyFilename)
	if err != nil {
		fmt.Printf("Error rendering daily archive file: %v\n", err)
	} else {
		fmt.Println("Daily archive file saved successfully:", dailyFilename)
	}

	err = GenerateDailyIndex()
	if err != nil {
		fmt.Printf("Error generating daily_trending/history.html: %v\n", err)
	} else {
		fmt.Println("Daily_trending/history.html updated successfully.")
	}

	err = generateLLMSJSON(todayStr, groupedItems)
	if err != nil {
		fmt.Printf("Error generating llms.json: %v\n", err)
	} else {
		fmt.Println("llms.json updated successfully.")
	}

	if os.Getenv("GITHUB_ACTIONS") != "true" {
		open("index.html")
	}
}
