package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

type ArchiveFile struct {
	Name        string
	DateDisplay string
	SizeDisplay string
	YearMonth   string
	IsLatest    bool
	parsedDate  time.Time
}

type ArchiveMonth struct {
	YearMonth   string
	DisplayName string
	Files       []ArchiveFile
	IsExpanded  bool
}

type HistoryPageData struct {
	TotalFiles  int
	StartDate   string
	EndDate     string
	LatestDate  string
	GeneratedAt string
	Months      []ArchiveMonth
}

func formatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%dB", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.1fKB", float64(size)/1024)
	}
	return fmt.Sprintf("%.1fMB", float64(size)/(1024*1024))
}

func generateFileListData() ([]ArchiveFile, error) {
	entries, err := os.ReadDir("daily_trending")
	if err != nil {
		return nil, fmt.Errorf("read daily_trending: %w", err)
	}

	files := make([]ArchiveFile, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".html") || name == "index.html" {
			continue
		}

		dateISO := strings.TrimSuffix(name, ".html")
		parsedDate, err := time.Parse("2006-01-02", dateISO)
		if err != nil {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, ArchiveFile{
			Name:        name,
			DateDisplay: parsedDate.Format("2006年1月2日"),
			SizeDisplay: formatFileSize(info.Size()),
			YearMonth:   parsedDate.Format("2006-01"),
			parsedDate:  parsedDate,
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].parsedDate.After(files[j].parsedDate)
	})

	if len(files) > 0 {
		files[0].IsLatest = true
	}

	return files, nil
}

func groupFilesByMonth(files []ArchiveFile) []ArchiveMonth {
	if len(files) == 0 {
		return nil
	}

	monthMap := make(map[string][]ArchiveFile)
	monthOrder := make([]string, 0, len(files))

	for _, file := range files {
		if _, exists := monthMap[file.YearMonth]; !exists {
			monthOrder = append(monthOrder, file.YearMonth)
		}
		monthMap[file.YearMonth] = append(monthMap[file.YearMonth], file)
	}

	groups := make([]ArchiveMonth, 0, len(monthOrder))
	for idx, ym := range monthOrder {
		displayName := ym
		if monthDate, err := time.Parse("2006-01", ym); err == nil {
			displayName = monthDate.Format("2006年1月")
		}

		groups = append(groups, ArchiveMonth{
			YearMonth:   ym,
			DisplayName: displayName,
			Files:       monthMap[ym],
			IsExpanded:  idx == 0,
		})
	}

	return groups
}

func GenerateDailyIndex() error {
	files, err := generateFileListData()
	if err != nil {
		return fmt.Errorf("build archive data: %w", err)
	}
	if len(files) == 0 {
		return fmt.Errorf("no HTML files found in daily_trending")
	}

	monthGroups := groupFilesByMonth(files)
	data := HistoryPageData{
		TotalFiles:  len(files),
		StartDate:   files[len(files)-1].DateDisplay,
		EndDate:     files[0].DateDisplay,
		LatestDate:  files[0].DateDisplay,
		GeneratedAt: time.Now().Format("2006-01-02"),
		Months:      monthGroups,
	}

	if err := renderTemplateToFile("templates/history.tmpl", data, "daily_trending/history.html"); err != nil {
		return fmt.Errorf("render archive template: %w", err)
	}
	if err := generateSitemap(files); err != nil {
		return fmt.Errorf("generate sitemap: %w", err)
	}

	return nil
}

func generateSitemap(files []ArchiveFile) error {
	baseURL := "https://vcvvvc.github.io/github_trending"
	currentTime := time.Now().Format("2006-01-02T15:04:05-07:00")

	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	<url>
		<loc>`)
	builder.WriteString(baseURL)
	builder.WriteString(`/</loc>
		<lastmod>`)
	builder.WriteString(currentTime)
	builder.WriteString(`</lastmod>
		<changefreq>daily</changefreq>
		<priority>1.0</priority>
	</url>
	<url>
		<loc>`)
	builder.WriteString(baseURL)
	builder.WriteString(`/daily_trending/</loc>
		<lastmod>`)
	builder.WriteString(currentTime)
	builder.WriteString(`</lastmod>
		<changefreq>daily</changefreq>
		<priority>0.9</priority>
	</url>`)

	for _, file := range files {
		builder.WriteString(`
	<url>
		<loc>`)
		builder.WriteString(baseURL)
		builder.WriteString(`/daily_trending/`)
		builder.WriteString(file.Name)
		builder.WriteString(`</loc>
		<lastmod>`)
		builder.WriteString(file.parsedDate.Format("2006-01-02T15:04:05-07:00"))
		builder.WriteString(`</lastmod>
		<changefreq>weekly</changefreq>
		<priority>0.8</priority>
	</url>`)
	}

	builder.WriteString(`
</urlset>`)

	return saveHTMLToFile("sitemap.xml", builder.String())
}
