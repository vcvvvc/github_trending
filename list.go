package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// 文件信息结构
type FileInfo struct {
	Name     string
	Date     string
	Size     string
	IsLatest bool
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
				Name:     entry.Name(),
				Date:     date.Format("2006年1月2日"),
				Size:     size,
				IsLatest: isLatest,
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

// 生成daily_trending/index.html文件
func GenerateDailyIndex() error {
	files, err := generateFileListData()
	if err != nil {
		return fmt.Errorf("Error generating file list data: %v", err)
	}
	
	if len(files) == 0 {
		return fmt.Errorf("No HTML files found in daily_trending directory")
	}
	
	// 计算统计信息
	startDate := strings.TrimSuffix(files[len(files)-1].Name, ".html")
	endDate := strings.TrimSuffix(files[0].Name, ".html")
	
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
            font-size: 1em;
            color: white;
            font-weight: 500;
            font-family: 'Inter', sans-serif;
            box-shadow: 0 4px 16px rgba(102, 126, 234, 0.3);
        }
        
        .stats span {
            display: flex;
            align-items: center;
            gap: 8px;
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
        

        
        .back-link {
            display: inline-block;
            margin-top: 20px;
            padding: 10px 20px;
            background: #0366d6;
            color: white;
            text-decoration: none;
            border-radius: 6px;
            transition: background-color 0.3s ease;
        }
        
        .back-link:hover {
            background: #0256cc;
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
        <span>📊 总计: %d 个历史文件</span>
        <span>📅 时间跨度: %s 至 %s</span>
    </div>
    
    <div class="file-list">`, len(files), startDate, endDate)
	
	// 添加文件列表
	for _, file := range files {
		latestClass := ""
		latestBadge := ""
		if file.IsLatest {
			latestClass = " latest"
			latestBadge = `<span class="latest-badge">最新</span>`
		}
		
		htmlContent += fmt.Sprintf(`
        <div class="file-item%s" onclick="window.open('%s', '_blank')">
            <div class="file-icon">📅</div>
            <div class="file-info">
                <div class="file-name">%s%s</div>
                <div class="file-date">%s</div>
            </div>
            <div class="file-size">%s</div>
        </div>`, latestClass, file.Name, file.Name, latestBadge, file.Date, file.Size)
	}
	
	htmlContent += `
    </div>
    
    <a href="../index.html" class="back-link">← 返回主页</a>
    
    <script>
        // 添加点击效果
        document.querySelectorAll('.file-item').forEach(item => {
            item.addEventListener('click', function() {
                this.style.transform = 'scale(0.98)';
                setTimeout(() => {
                    this.style.transform = '';
                }, 150);
            });
        });
        
        // 页面加载动画
        document.addEventListener('DOMContentLoaded', function() {
            const items = document.querySelectorAll('.file-item');
            items.forEach((item, index) => {
                item.style.opacity = '0';
                item.style.transform = 'translateY(20px)';
                setTimeout(() => {
                    item.style.transition = 'all 0.5s ease';
                    item.style.opacity = '1';
                    item.style.transform = 'translateY(0)';
                }, index * 100);
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