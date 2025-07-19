package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

var calendarSVG = `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" style="width:1.2em;height:1.2em;vertical-align:-0.2em;"><path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5"/></svg>`
var hourglassSVG = `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" style="width:1.2em;height:1.2em;vertical-align:-0.2em;"><path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3h10.5M6.75 21h10.5M6.75 3v3.375c0 1.192.464 2.335 1.293 3.182l2.457 2.4a3.75 3.75 0 0 1 0 5.086l-2.457 2.4A4.5 4.5 0 0 0 6.75 17.625V21m10.5-18v3.375c0 1.192-.464 2.335-1.293 3.182l-2.457 2.4a3.75 3.75 0 0 0 0 5.086l2.457 2.4c.829.808 1.293 1.95 1.293 3.182V21"/></svg>`

// 文件信息结构
type FileInfo struct {
	Name     string
	Date     string
	Size     string
	IsLatest bool
	YearMonth string
}

// 月份分组结构
type MonthGroup struct {
	YearMonth string
	DisplayName string
	Files     []FileInfo
	IsExpanded bool
}

// 生成文件列表数据
func generateFileListData() ([]FileInfo, error) {
	var files []FileInfo
	
	// 读取daily_trending目录中的所有html文件
	entries, err := os.ReadDir("daily_trending")
	if err != nil {
		return nil, fmt.Errorf("Error reading daily_trending directory: %v", err)
	}
	
	// 过滤出html文件并收集信息
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".html") && entry.Name() != "index.html" {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			
			// 解析日期
			dateStr := strings.TrimSuffix(entry.Name(), ".html")
			date, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				continue
			}
			
			// 格式化文件大小
			size := fmt.Sprintf("%.0fKB", float64(info.Size())/1024)
			
			// 检查是否是最新的文件
			isLatest := date.Format("2006-01-02") == time.Now().Format("2006-01-02")
			
			files = append(files, FileInfo{
				Name:      entry.Name(),
				Date:      date.Format("2006年1月2日"),
				Size:      size,
				IsLatest:  isLatest,
				YearMonth: date.Format("2006-01"),
			})
		}
	}
	
	// 按日期倒序排序（最新的在前）
	sort.Slice(files, func(i, j int) bool {
		dateI, _ := time.Parse("2006-01-02", strings.TrimSuffix(files[i].Name, ".html"))
		dateJ, _ := time.Parse("2006-01-02", strings.TrimSuffix(files[j].Name, ".html"))
		return dateI.After(dateJ)
	})
	
	return files, nil
}

// 按月份分组文件
func groupFilesByMonth(files []FileInfo) []MonthGroup {
	monthMap := make(map[string][]FileInfo)
	
	// 按月份分组
	for _, file := range files {
		monthMap[file.YearMonth] = append(monthMap[file.YearMonth], file)
	}
	
	// 转换为MonthGroup切片
	var groups []MonthGroup
	for yearMonth, monthFiles := range monthMap {
		// 解析年月来生成显示名称
		date, _ := time.Parse("2006-01", yearMonth)
		displayName := date.Format("2006年1月")
		
		// 最新月份默认展开
		isExpanded := yearMonth == files[0].YearMonth
		
		groups = append(groups, MonthGroup{
			YearMonth:  yearMonth,
			DisplayName: displayName,
			Files:      monthFiles,
			IsExpanded: isExpanded,
		})
	}
	
	// 按年月倒序排序
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].YearMonth > groups[j].YearMonth
	})
	
	return groups
}

// 生成daily_trending/index.html文件
func GenerateDailyIndex() error {
	files, err := generateFileListData()
	if err != nil {
		return fmt.Errorf("Error generating file list data: %v", err)
	}
	
	if len(files) == 0 {
		return fmt.Errorf("No HTML files found in daily_trending directory")
	}
	
	// 按月份分组
	monthGroups := groupFilesByMonth(files)
	
	// 计算统计信息
	startDate := strings.TrimSuffix(files[len(files)-1].Name, ".html")
	endDate := strings.TrimSuffix(files[0].Name, ".html")
	
	// 格式化日期显示
	startDateFormatted := startDate
	endDateFormatted := endDate
	
	// 尝试解析并格式化日期
	if startDateParsed, err := time.Parse("2006-01-02", startDate); err == nil {
		startDateFormatted = startDateParsed.Format("2006年1月2日")
	}
	if endDateParsed, err := time.Parse("2006-01-02", endDate); err == nil {
		endDateFormatted = endDateParsed.Format("2006年1月2日")
	}
	
	// 生成HTML内容
	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GitHub Trending 历史记录</title>
    <style>
        body {
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f6f8fa;
            color: #24292e;
        }
        
        @import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800&display=swap');
        
        .header {
            text-align: center;
            margin-bottom: 40px;
            padding: 30px;
            background: linear-gradient(135deg, #1a1a2e 0%%, #16213e 50%%, #0f3460 100%%);
            color: white;
            border-radius: 16px;
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
            position: relative;
            overflow: hidden;
        }
        
        .back-home-link {
            position: fixed;
            bottom: 100px;
            right: 30px;
            background: linear-gradient(145deg, #0366d6, #0256cc);
            color: white;
            border: 1px solid rgba(255, 255, 255, 0.2);
            border-radius: 12px;
            width: 48px;
            height: 48px;
            display: flex;
            align-items: center;
            justify-content: center;
            cursor: pointer;
            text-decoration: none;
            font-size: 20px;
            transition: all 0.3s ease-out;
            box-shadow: 0 6px 15px rgba(0, 0, 0, 0.2);
            z-index: 1000;
        }

        .back-home-link:hover {
            transform: translateY(-3px);
            box-shadow: 0 8px 20px rgba(0, 0, 0, 0.25);
            background: linear-gradient(145deg, #0256cc, #0147a3);
        }
        
        .header::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: linear-gradient(45deg, rgba(255,255,255,0.1) 0%%, rgba(255,255,255,0.05) 100%%);
            pointer-events: none;
        }
        
        .header h1 {
            margin: 0;
            font-size: 3em;
            font-weight: 800;
            font-family: 'Inter', sans-serif;
            color: #ffffff;
            text-shadow: 
                0 0 20px rgba(255, 255, 255, 0.3),
                0 4px 8px rgba(0, 0, 0, 0.4);
            letter-spacing: -0.02em;
            padding: 10px 0;
            position: relative;
            z-index: 1;
        }
        
        .header p {
            margin: 15px 0 0 0;
            font-size: 1.2em;
            font-weight: 400;
            font-family: 'Inter', sans-serif;
            color: rgba(255, 255, 255, 0.95);
            text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
            position: relative;
            z-index: 1;
        }
        
        .stats {
            display: flex;
            justify-content: space-between;
            margin: 20px 0 30px 0;
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            border-radius: 12px;
            font-size: 0.95em;
            color: white;
            font-weight: 500;
            font-family: 'Inter', sans-serif;
            box-shadow: 0 4px 16px rgba(102, 126, 234, 0.3);
            flex-wrap: wrap;
            gap: 15px;
        }
        
        .stats span {
            display: flex;
            align-items: center;
            gap: 8px;
            white-space: nowrap;
        }
        
        @media (max-width: 768px) {
            .stats {
                flex-direction: column;
                align-items: center;
                text-align: center;
            }
        }
        
        .counter {
            display: inline-block;
            font-weight: 700;
            color: #ffffff;
            text-shadow: 0 0 10px rgba(255, 255, 255, 0.5);
            transition: all 0.3s ease;
        }
        
        .counter.animate {
            transform: scale(1.1);
            color: #ffd700;
        }
        
        .view-toggle {
            display: flex;
            justify-content: center;
            margin: 20px 0;
            gap: 10px;
        }
        
        .toggle-btn {
            padding: 10px 20px;
            border: 2px solid #0366d6;
            background: white;
            color: #0366d6;
            border-radius: 25px;
            cursor: pointer;
            font-weight: 500;
            font-family: 'Inter', sans-serif;
            transition: all 0.3s ease;
            display: flex;
            align-items: center;
            gap: 8px;
        }
        
        .toggle-btn:hover {
            background: #0366d6;
            color: white;
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(3, 102, 214, 0.3);
        }
        
        .toggle-btn.active {
            background: #0366d6;
            color: white;
        }
        
        .timeline-view {
            display: none;
        }
        
        .timeline-view.active {
            display: block;
        }
        
        .month-view {
            display: block;
        }
        
        .month-view.hidden {
            display: none;
        }
        
        .vertical-timeline {
            position: relative;
            padding: 20px 0;
            max-height: 70vh;
            overflow-y: auto;
            scrollbar-width: none;
            -ms-overflow-style: none;
        }
        
        .vertical-timeline::-webkit-scrollbar {
            display: none;
        }
        
        .vertical-timeline::before {
            content: '';
            position: absolute;
            left: 50px;
            top: 0;
            bottom: 0;
            width: 3px;
            background: linear-gradient(to bottom, #0366d6, #28a745);
            border-radius: 2px;
        }
        
        .timeline-card {
            position: relative;
            margin: 20px 0;
            margin-left: 80px;
            background: white;
            border-radius: 12px;
            padding: 20px;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
            transition: all 0.3s ease;
            cursor: pointer;
            border-left: 4px solid #0366d6;
        }
        
        .timeline-card:hover {
            transform: translateY(-4px);
            box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15);
            border-left-color: #28a745;
        }
        
        .timeline-card.latest {
            border-left-color: #28a745;
            background: linear-gradient(135deg, #f8fff9 0%%, #e8f5e8 100%%);
            box-shadow: 0 6px 20px rgba(40, 167, 69, 0.2);
        }
        
        .timeline-card::before {
            content: '';
            position: absolute;
            left: -35px;
            top: 25px;
            width: 12px;
            height: 12px;
            background: #0366d6;
            border: 3px solid white;
            border-radius: 50%;
            box-shadow: 0 0 0 3px #0366d6;
        }
        
        .timeline-card.latest::before {
            background: #28a745;
            box-shadow: 0 0 0 3px #28a745;
        }
        
        .timeline-card::after {
            content: '';
            position: absolute;
            left: -25px;
            top: 31px;
            width: 0;
            height: 0;
            border-top: 6px solid transparent;
            border-bottom: 6px solid transparent;
            border-right: 8px solid white;
        }
        
        .timeline-card.latest::after {
            border-right-color: #f8fff9;
        }
        
        .timeline-date {
            font-size: 0.9em;
            color: #586069;
            margin-bottom: 8px;
            font-weight: 500;
        }
        
        .timeline-title {
            font-size: 1.1em;
            font-weight: 600;
            color: #0366d6;
            margin-bottom: 5px;
        }
        
        .timeline-card.latest .timeline-title {
            color: #28a745;
        }
        
        .timeline-size {
            font-size: 0.8em;
            color: #586069;
            margin-top: 5px;
        }
        
        .latest-badge {
            background: #28a745;
            color: white;
            padding: 2px 8px;
            border-radius: 12px;
            font-size: 0.8em;
            margin-left: 10px;
        }
        
        .month-group {
            margin-bottom: 20px;
            border: 1px solid #e1e4e8;
            border-radius: 8px;
            overflow: hidden;
            background: white;
        }
        
        .month-header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            padding: 15px 20px;
            background: linear-gradient(135deg, #f8f9fa 0%%, #e9ecef 100%%);
            cursor: pointer;
            transition: all 0.3s ease;
            border-bottom: 1px solid #e1e4e8;
        }
        
        .month-header:hover {
            background: linear-gradient(135deg, #e9ecef 0%%, #dee2e6 100%%);
        }
        
        .month-title {
            font-weight: 600;
            font-size: 1.1em;
            color: #24292e;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        
        .month-count {
            background: #0366d6;
            color: white;
            padding: 2px 8px;
            border-radius: 12px;
            font-size: 0.8em;
            font-weight: 500;
        }
        
        .month-toggle {
            font-size: 1.2em;
            color: #586069;
            transition: transform 0.3s ease;
        }
        
        .month-toggle.expanded {
            transform: rotate(180deg);
        }
        
        .month-content {
            max-height: 0;
            overflow: hidden;
            transition: max-height 0.3s ease;
        }
        
        .month-content.expanded {
            max-height: 2000px;
        }
        
        .month-files {
            padding: 0;
        }
        
        .file-list {
            background: white;
            border-radius: 10px;
            padding: 30px;
            box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
        }
        
        .file-item {
            display: flex;
            align-items: center;
            padding: 15px;
            margin: 10px 0;
            border: 1px solid #e1e4e8;
            border-radius: 6px;
            transition: all 0.3s ease;
            text-decoration: none;
            color: inherit;
            cursor: pointer;
        }
        
        .file-item:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
            border-color: #0366d6;
        }
        
        .file-item.latest {
            border-color: #28a745;
            background-color: #f8fff9;
        }
        
        .file-icon {
            font-size: 1.5em;
            margin-right: 15px;
            color: #0366d6;
        }
        
        .file-item.latest .file-icon {
            color: #28a745;
        }
        
        .file-info {
            flex: 1;
        }
        
        .file-name {
            font-weight: 600;
            font-size: 1.1em;
            color: #0366d6;
            margin-bottom: 5px;
        }
        
        .file-item.latest .file-name {
            color: #28a745;
        }
        
        .file-date {
            color: #586069;
            font-size: 0.9em;
        }
        
        .file-size {
            color: #586069;
            font-size: 0.9em;
            margin-left: 10px;
        }
        
        .latest-badge {
            background: #28a745;
            color: white;
            padding: 2px 8px;
            border-radius: 12px;
            font-size: 0.8em;
            margin-left: 10px;
        }
        
        .calendar-svg {
            display: inline-block;
            vertical-align: middle;
            margin-right: 4px;
        }
        

        

        
        @media (max-width: 600px) {
            body {
                padding: 10px;
            }
            
            .header h1 {
                font-size: 2em;
            }
            
            .file-list {
                padding: 20px;
            }
            
            .file-item {
                flex-direction: column;
                align-items: flex-start;
            }
            
            .file-icon {
                margin-bottom: 10px;
            }
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>📊 GitHub Trending 历史记录</h1>
        <p>查看所有历史趋势数据</p>
    </div>
    
    <div class="stats">
        <span>📊 总计: <span class="counter" data-target="` + fmt.Sprintf("%d", len(files)) + `">0</span> 个历史文件</span>
        <span>📅 时间跨度: ` + startDateFormatted + ` 至 ` + endDateFormatted + `</span>
        <span>🔄 最近更新: ` + files[0].Date + `</span>
    </div>
    
    <div class="view-toggle">
        <button class="toggle-btn active" onclick="switchView('month')">
            <span class="calendar-svg">` + calendarSVG + `</span> 月卡片盒
        </button>
        <button class="toggle-btn" onclick="switchView('timeline')">
            <span class="calendar-svg">` + hourglassSVG + `</span> 时间线
        </button>
    </div>
    
    <div class="month-view" id="month-view">
        <div class="file-list">`)
	
	// 添加月份分组
	for _, group := range monthGroups {
		expandedClass := ""
		if group.IsExpanded {
			expandedClass = " expanded"
		}
		
		htmlContent += fmt.Sprintf(`
        <div class="month-group">
            <div class="month-header" onclick="toggleMonth('%s')">
                <div class="month-title">
                    <span class="calendar-svg">%s</span> %s
                    <span class="month-count">%d</span>
                </div>
                <div class="month-toggle%s">▼</div>
            </div>
            <div class="month-content%s">
                <div class="month-files">`, group.YearMonth, calendarSVG, group.DisplayName, len(group.Files), expandedClass, expandedClass)
		
		// 添加该月份的文件
		for _, file := range group.Files {
			latestClass := ""
			latestBadge := ""
			if file.IsLatest {
				latestClass = " latest"
				latestBadge = `<span class="latest-badge">最新</span>`
			}
			
			htmlContent += fmt.Sprintf(`
                <div class="file-item%s" onclick="window.open('%s', '_blank')">
                    <div class="file-icon"><span class="calendar-svg">%s</span></div>
                    <div class="file-info">
                        <div class="file-name">%s%s</div>
                        <div class="file-date">%s</div>
                    </div>
                    <div class="file-size">%s</div>
                </div>`, latestClass, file.Name, calendarSVG, file.Name, latestBadge, file.Date, file.Size)
		}
		
		htmlContent += `
                </div>
            </div>
        </div>`
	}
	
	htmlContent += `
    </div>
    </div>
    
    <div class="timeline-view" id="timeline-view">
        <div class="vertical-timeline">`
	
	// 添加时间线卡片
	for _, file := range files {
		latestClass := ""
		latestBadge := ""
		if file.IsLatest {
			latestClass = " latest"
			latestBadge = `<span class="latest-badge">最新</span>`
		}
		
		htmlContent += fmt.Sprintf(`
            <div class="timeline-card%s" onclick="window.open('%s', '_blank')">
                <div class="timeline-date">%s</div>
                <div class="timeline-title">%s%s</div>
                <div class="timeline-size">%s</div>
            </div>`, latestClass, file.Name, file.Date, file.Name, latestBadge, file.Size)
	}
	
	htmlContent += `
        </div>
    </div>
    
    <a href="../index.html" class="back-home-link" title="返回主页">🏠</a>
    
    <script>
        // 数字增长动画函数
        function animateCounter(element, target, duration = 2000) {
            const start = 0;
            const increment = target / (duration / 16); // 60fps
            let current = start;
            
            const timer = setInterval(() => {
                current += increment;
                if (current >= target) {
                    current = target;
                    clearInterval(timer);
                    element.classList.add('animate');
                    setTimeout(() => {
                        element.classList.remove('animate');
                    }, 500);
                }
                element.textContent = Math.floor(current);
            }, 16);
        }
        
        // 视图切换函数
        function switchView(viewType) {
            const monthView = document.getElementById('month-view');
            const timelineView = document.getElementById('timeline-view');
            const monthBtn = document.querySelector('[onclick="switchView(\'month\')"]');
            const timelineBtn = document.querySelector('[onclick="switchView(\'timeline\')"]');
            
            if (viewType === 'month') {
                monthView.classList.remove('hidden');
                timelineView.classList.remove('active');
                monthBtn.classList.add('active');
                timelineBtn.classList.remove('active');
            } else {
                monthView.classList.add('hidden');
                timelineView.classList.add('active');
                monthBtn.classList.remove('active');
                timelineBtn.classList.add('active');
            }
        }
        
        // 月份折叠切换函数
        function toggleMonth(yearMonth) {
            const monthGroup = document.querySelector('[onclick="toggleMonth(\'' + yearMonth + '\')"]').closest('.month-group');
            const content = monthGroup.querySelector('.month-content');
            const toggle = monthGroup.querySelector('.month-toggle');
            
            if (content.classList.contains('expanded')) {
                content.classList.remove('expanded');
                toggle.classList.remove('expanded');
            } else {
                content.classList.add('expanded');
                toggle.classList.add('expanded');
            }
        }
        
        // 添加点击效果
        document.querySelectorAll('.file-item, .timeline-card').forEach(item => {
            item.addEventListener('click', function() {
                this.style.transform = 'scale(0.98)';
                setTimeout(() => {
                    this.style.transform = '';
                }, 150);
            });
        });
        
        // 页面加载动画
        document.addEventListener('DOMContentLoaded', function() {
            // 启动数字增长动画
            const counter = document.querySelector('.counter');
            if (counter) {
                const target = parseInt(counter.getAttribute('data-target'));
                setTimeout(() => {
                    animateCounter(counter, target, 1500);
                }, 500);
            }
            
            // 月份组动画
            const monthGroups = document.querySelectorAll('.month-group');
            monthGroups.forEach((group, groupIndex) => {
                group.style.opacity = '0';
                group.style.transform = 'translateY(20px)';
                setTimeout(() => {
                    group.style.transition = 'all 0.5s ease';
                    group.style.opacity = '1';
                    group.style.transform = 'translateY(0)';
                }, groupIndex * 200 + 800);
                
                // 文件项动画
                const items = group.querySelectorAll('.file-item');
                items.forEach((item, index) => {
                    item.style.opacity = '0';
                    item.style.transform = 'translateY(10px)';
                    setTimeout(() => {
                        item.style.transition = 'all 0.3s ease';
                        item.style.opacity = '1';
                        item.style.transform = 'translateY(0)';
                    }, groupIndex * 200 + 1000 + index * 50);
                });
            });
        });
    </script>
</body>
</html>`
	
	// 保存到文件
	indexFilename := "daily_trending/index.html"
	err = saveHTMLToFile(indexFilename, htmlContent)
	if err != nil {
		return fmt.Errorf("Error saving daily_trending/index.html: %v", err)
	}
	
	return nil
} 