package main

import (
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/yldm-tech/mygo-kubai-icons/all"
)

// words is what the gallery says, in one language.
type words struct {
	zh bool

	title, tagline, install, copy, copied, github, docs, upstream, otherLang string

	all, search, look, color, soft, size, basesOnly, category string
	accents                                                   []string

	hint, count, copiedCall, empty string

	pick, preview, previewHint, active, name, function, inCategory, code, copyCode, variants string

	credit, mit, unofficial, drawnBy string
}

var english = words{
	title:     "Kubai Icons for MyGo",
	tagline:   "2,871 icons that move on hover, drawn by MyGo native UI",
	install:   "go get github.com/yldm-tech/mygo-kubai-icons",
	copy:      "Copy",
	copied:    "Copied",
	github:    "GitHub",
	docs:      "Docs",
	upstream:  "Kubai Icons",
	otherLang: "中文",

	all: "All icons", search: "Search icons", look: "Look", color: "Color", soft: "Soft", size: "Size", basesOnly: "Base icons only", category: "Category",
	accents: []string{"Blue", "Violet", "Pink", "Orange", "Green", "Slate"},

	hint:       "Hover an icon to play it; click it for its code.",
	count:      "%d of %d icons",
	copiedCall: "Copied %s",
	empty:      "No icon matches. Try another word, or all categories.",

	pick:        "Pick an icon to see it large, with its code and its other versions.",
	preview:     "Preview",
	previewHint: "Hover to play",
	active:      "Toggled on",
	name:        "Name",
	function:    "Function",
	inCategory:  "Category",
	code:        "Code",
	copyCode:    "Copy code",
	variants:    "Versions",

	credit:     "Icons from",
	mit:        "MIT",
	unofficial: "Unofficial port",
	drawnBy:    "Drawn by MyGo, compiled to WebAssembly",
}

var chinese = words{
	zh:        true,
	title:     "Kubai Icons for MyGo",
	tagline:   "2,871 个悬停会动的图标，由 MyGo 原生 UI 绘制",
	install:   "go get github.com/yldm-tech/mygo-kubai-icons",
	copy:      "复制",
	copied:    "已复制",
	github:    "GitHub",
	docs:      "文档",
	upstream:  "Kubai Icons",
	otherLang: "English",

	all: "全部图标", search: "搜索图标", look: "风格", color: "彩色", soft: "清透", size: "尺寸", basesOnly: "只看基础图标", category: "分类",
	accents: []string{"蓝", "紫", "粉", "橙", "绿", "灰"},

	hint:       "悬停播放动画，点击查看代码。",
	count:      "%d / %d 个图标",
	copiedCall: "已复制 %s",
	empty:      "没有匹配的图标。换个词试试，或者看全部分类。",

	pick:        "选一个图标，这里会显示大图、代码和同组的其他版本。",
	preview:     "预览",
	previewHint: "悬停播放",
	active:      "选中状态",
	name:        "名称",
	function:    "函数",
	inCategory:  "分类",
	code:        "代码",
	copyCode:    "复制代码",
	variants:    "同组版本",

	credit:     "图标来自",
	mit:        "MIT 协议",
	unofficial: "非官方移植",
	drawnBy:    "由 MyGo 绘制，编译为 WebAssembly",
}

// say picks the language the app shows: Chinese for any zh locale, English otherwise.
func say(c *ui.Context) words {
	if strings.HasPrefix(c.Locale(), "zh") {
		return chinese
	}
	return english
}

// label is an icon's name as the gallery shows it: upstream's Chinese name in Chinese, its id otherwise.
func (w words) label(e all.Entry) string {
	if w.zh && e.Zh != "" {
		return e.Zh
	}
	return e.Name
}

// categoryName is a category's name in the gallery's language.
func (w words) categoryName(id string) string {
	for _, cat := range all.Categories {
		if cat.ID == id {
			if w.zh {
				return cat.Zh
			}
			return cat.En
		}
	}
	return id
}
