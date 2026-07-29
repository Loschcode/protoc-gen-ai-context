package generator

import (
	"fmt"
	"sort"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/Loschcode/protoc-gen-ai-context/internal/annotations"
)

// TopicFile holds all collected data for one knowledge topic.
type TopicFile struct {
	Topic    string
	Title    string
	Sections []Section
	// Aggregated from all annotations belonging to this topic.
	summaries     []string
	humanNames    []string
	commonQueries []string
	usageNotes    []string
}

// Section is one block in the generated markdown (enum, message, etc.)
type Section struct {
	Heading string
	Body    string
}

// Collector gathers annotations across proto files and groups them by topic.
type Collector struct {
	topics map[string]*TopicFile
}

// NewCollector creates a new annotation collector.
func NewCollector() *Collector {
	return &Collector{topics: make(map[string]*TopicFile)}
}

func (c *Collector) getOrCreate(topic, title string) *TopicFile {
	tf, ok := c.topics[topic]
	if !ok {
		tf = &TopicFile{Topic: topic, Title: title}
		c.topics[topic] = tf
	}
	if tf.Title == "" && title != "" {
		tf.Title = title
	}
	return tf
}

// CollectFile processes a single proto file for knowledge annotations.
func (c *Collector) CollectFile(file *protogen.File) {
	desc := file.Desc

	// Walk enums
	for i := 0; i < desc.Enums().Len(); i++ {
		c.collectEnum(desc.Enums().Get(i))
	}

	// Walk messages (and their nested enums/messages)
	for i := 0; i < desc.Messages().Len(); i++ {
		c.collectMessage(desc.Messages().Get(i))
	}
}

func (c *Collector) collectEnum(enum protoreflect.EnumDescriptor) {
	kt, hasKT := annotations.KnowledgeFromEnum(enum)
	parentTopic := ""
	if hasKT {
		parentTopic = kt.Topic
		tf := c.getOrCreate(kt.Topic, kt.Title)
		if kt.Summary != "" {
			tf.summaries = append(tf.summaries, kt.Summary)
		}
		if kt.UsageNotes != "" {
			tf.usageNotes = append(tf.usageNotes, kt.UsageNotes)
		}
	}

	// Collect enum values
	var valueLines []string
	for i := 0; i < enum.Values().Len(); i++ {
		val := enum.Values().Get(i)
		ctx, hasCtx := annotations.ContextFromEnumValue(val)
		if !hasCtx {
			continue
		}

		topic := parentTopic
		if ctx.Topic != "" {
			topic = ctx.Topic
		}
		if topic == "" {
			continue
		}

		tf := c.getOrCreate(topic, "")
		if ctx.HumanName != "" {
			tf.humanNames = append(tf.humanNames, ctx.HumanName)
		}
		if ctx.CommonQueries != "" {
			tf.commonQueries = append(tf.commonQueries, ctx.CommonQueries)
		}

		// Build a line for the values section
		name := string(val.Name())
		comment := extractLeadingComment(val)
		line := fmt.Sprintf("- **%s** (`%s`)", ctx.HumanName, name)
		if comment != "" {
			line += " — " + comment
		}
		if ctx.CommonQueries != "" {
			line += fmt.Sprintf(" (users may say: %s)", ctx.CommonQueries)
		}
		valueLines = append(valueLines, line)
	}

	if len(valueLines) > 0 && parentTopic != "" {
		tf := c.getOrCreate(parentTopic, "")
		enumName := string(enum.Name())
		comment := extractLeadingComment(enum)
		heading := enumName
		if comment != "" {
			heading = comment
		}
		tf.Sections = append(tf.Sections, Section{
			Heading: heading,
			Body:    strings.Join(valueLines, "\n"),
		})
	}
}

func (c *Collector) collectMessage(msg protoreflect.MessageDescriptor) {
	kt, hasKT := annotations.KnowledgeFromMessage(msg)
	if hasKT {
		tf := c.getOrCreate(kt.Topic, kt.Title)
		if kt.Summary != "" {
			tf.summaries = append(tf.summaries, kt.Summary)
		}
		if kt.UsageNotes != "" {
			tf.usageNotes = append(tf.usageNotes, kt.UsageNotes)
		}

		// Build field documentation
		var fieldLines []string
		for i := 0; i < msg.Fields().Len(); i++ {
			field := msg.Fields().Get(i)
			fieldLine := formatField(field)
			fieldLines = append(fieldLines, fieldLine)
		}

		body := ""
		if len(fieldLines) > 0 {
			body += "Fields:\n" + strings.Join(fieldLines, "\n")
		}
		if kt.Example != "" {
			body += "\n\nExample payload:\n```json\n" + kt.Example + "\n```"
		}

		msgName := string(msg.Name())
		comment := extractLeadingComment(msg)
		heading := msgName
		if comment != "" {
			heading = comment + " (" + msgName + ")"
		}

		tf.Sections = append(tf.Sections, Section{
			Heading: heading,
			Body:    body,
		})
	}

	// Recurse into nested enums and messages
	for i := 0; i < msg.Enums().Len(); i++ {
		c.collectEnum(msg.Enums().Get(i))
	}
	for i := 0; i < msg.Messages().Len(); i++ {
		c.collectMessage(msg.Messages().Get(i))
	}
}

func formatField(field protoreflect.FieldDescriptor) string {
	name := string(field.Name())
	typeName := fieldTypeName(field)
	required := ""
	comment := extractLeadingComment(field)

	// Check for FieldContext annotation
	fctx, hasCtx := annotations.ContextFromField(field)
	if hasCtx && fctx.Description != "" {
		comment = fctx.Description
	}

	example := ""
	if hasCtx && fctx.Example != "" {
		example = fmt.Sprintf(" (e.g. `%s`)", fctx.Example)
	}

	if field.HasOptionalKeyword() {
		required = " *(optional)*"
	}

	line := fmt.Sprintf("- `%s` (%s)%s", name, typeName, required)
	if comment != "" {
		line += " — " + comment
	}
	line += example
	return line
}

func fieldTypeName(field protoreflect.FieldDescriptor) string {
	if field.IsList() {
		return "repeated " + scalarOrRef(field)
	}
	if field.IsMap() {
		return "map"
	}
	return scalarOrRef(field)
}

func scalarOrRef(field protoreflect.FieldDescriptor) string {
	switch field.Kind() {
	case protoreflect.MessageKind:
		return string(field.Message().Name())
	case protoreflect.EnumKind:
		return string(field.Enum().Name())
	default:
		return field.Kind().String()
	}
}

// Generate produces markdown content for all collected topics.
func (c *Collector) Generate() map[string]string {
	result := make(map[string]string)

	// Sort topics for deterministic output
	var topicNames []string
	for name := range c.topics {
		topicNames = append(topicNames, name)
	}
	sort.Strings(topicNames)

	for _, name := range topicNames {
		tf := c.topics[name]
		result[name] = tf.render()
	}
	return result
}

func (tf *TopicFile) render() string {
	var b strings.Builder

	// Heading
	title := tf.Title
	if title == "" {
		title = tf.Topic
	}
	b.WriteString("# ")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Build the capability summary (first paragraph — critical for topic index)
	summary := tf.buildSummary()
	if summary != "" {
		b.WriteString(summary)
		b.WriteString("\n")
	}

	// Usage notes (behavioral rules not derivable from schema)
	if len(tf.usageNotes) > 0 {
		b.WriteString("\n## Usage Notes\n\n")
		for _, note := range tf.usageNotes {
			b.WriteString(note)
			b.WriteString("\n\n")
		}
	}

	// Sections
	for _, s := range tf.Sections {
		b.WriteString("\n## ")
		b.WriteString(s.Heading)
		b.WriteString("\n\n")
		b.WriteString(s.Body)
		b.WriteString("\n")
	}

	return b.String()
}

func (tf *TopicFile) buildSummary() string {
	var parts []string

	// Add all summaries
	for _, s := range tf.summaries {
		parts = append(parts, s)
	}

	// Add human names as a capability list
	if len(tf.humanNames) > 0 {
		unique := uniqueStrings(tf.humanNames)
		parts = append(parts, "Capabilities: "+strings.Join(unique, ", ")+".")
	}

	// Add common queries
	if len(tf.commonQueries) > 0 {
		unique := uniqueStrings(tf.commonQueries)
		parts = append(parts, "Common queries: "+strings.Join(unique, ", ")+".")
	}

	return strings.Join(parts, " ")
}

func uniqueStrings(input []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range input {
		lower := strings.ToLower(strings.TrimSpace(s))
		if !seen[lower] && s != "" {
			seen[lower] = true
			result = append(result, strings.TrimSpace(s))
		}
	}
	return result
}

// extractLeadingComment gets the proto comment for a descriptor.
func extractLeadingComment(desc protoreflect.Descriptor) string {
	si := desc.ParentFile().SourceLocations().ByDescriptor(desc)
	comment := strings.TrimSpace(si.LeadingComments)
	if comment == "" {
		return ""
	}
	// Take the first sentence/line
	lines := strings.SplitN(comment, "\n", 2)
	return strings.TrimSpace(lines[0])
}
