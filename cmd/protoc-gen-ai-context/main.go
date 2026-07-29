package main

import (
	"strings"

	"github.com/Loschcode/protoc-gen-ai-context/internal/generator"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	opts := protogen.Options{}
	opts.Run(func(plugin *protogen.Plugin) error {
		plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

		collector := generator.NewCollector()

		for _, file := range plugin.Files {
			if !file.Generate {
				continue
			}
			collector.CollectFile(file)
		}

		// Generate one .md file per topic
		topics := collector.Generate()
		if len(topics) == 0 {
			return nil
		}

		// Output files directly in the output root (no proto package subdirectory).
		for topic, content := range topics {
			filename := topic + ".ai.md"
			g := plugin.NewGeneratedFile(filename, "")
			g.P(strings.TrimRight(content, "\n"))
		}

		// Generate a summary index file
		indexContent := buildIndex(topics)
		g := plugin.NewGeneratedFile("_index.ai.md", "")
		g.P(strings.TrimRight(indexContent, "\n"))

		return nil
	})
}

// buildIndex creates a summary file listing all topics with their first paragraphs.
func buildIndex(topics map[string]string) string {
	var b strings.Builder
	b.WriteString("# AI Knowledge Index\n\n")
	b.WriteString("Auto-generated topic index. Each topic can be loaded on demand via `lookup_knowledge(\"<topic>\")` to get full details.\n\n")

	for topic, content := range topics {
		title, summary := extractHeader(content)
		b.WriteString("## ")
		b.WriteString(title)
		b.WriteString("\n\n")
		b.WriteString("**Topic:** `")
		b.WriteString(topic)
		b.WriteString("`\n\n")
		if summary != "" {
			b.WriteString(summary)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	return b.String()
}

// extractHeader pulls the first heading and first paragraph from markdown content.
func extractHeader(content string) (title, summary string) {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") && title == "" {
			title = strings.TrimPrefix(trimmed, "# ")
			for j := i + 1; j < len(lines); j++ {
				para := strings.TrimSpace(lines[j])
				if para == "" {
					continue
				}
				if strings.HasPrefix(para, "#") {
					break
				}
				summary = para
				break
			}
			break
		}
	}
	if title == "" {
		title = "Untitled"
	}
	return
}
