package annotations

import (
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// Field numbers matching annotations.proto extensions.
const (
	enumKnowledgeField    protowire.Number = 52100
	enumValueContextField protowire.Number = 52101
	messageKnowledgeField protowire.Number = 52102
	fieldContextField     protowire.Number = 52103
)

// KnowledgeTopic mirrors the proto KnowledgeTopic message.
type KnowledgeTopic struct {
	Topic   string
	Title   string
	Summary string
	Example string
}

// EnumValueContext mirrors the proto EnumValueContext message.
type EnumValueContext struct {
	HumanName     string
	CommonQueries string
	Topic         string
}

// FieldContext mirrors the proto FieldContext message.
type FieldContext struct {
	Description string
	Example     string
}

// KnowledgeFromEnum extracts the KnowledgeTopic annotation from an enum.
func KnowledgeFromEnum(enum protoreflect.EnumDescriptor) (KnowledgeTopic, bool) {
	opts, ok := enum.Options().(*descriptorpb.EnumOptions)
	if !ok || opts == nil {
		return KnowledgeTopic{}, false
	}
	raw := opts.ProtoReflect().GetUnknown()
	ext := findExtension(raw, enumKnowledgeField)
	if ext == nil {
		return KnowledgeTopic{}, false
	}
	return parseKnowledgeTopic(ext), true
}

// ContextFromEnumValue extracts the EnumValueContext annotation from an enum value.
func ContextFromEnumValue(value protoreflect.EnumValueDescriptor) (EnumValueContext, bool) {
	opts, ok := value.Options().(*descriptorpb.EnumValueOptions)
	if !ok || opts == nil {
		return EnumValueContext{}, false
	}
	raw := opts.ProtoReflect().GetUnknown()
	ext := findExtension(raw, enumValueContextField)
	if ext == nil {
		return EnumValueContext{}, false
	}
	return parseEnumValueContext(ext), true
}

// KnowledgeFromMessage extracts the KnowledgeTopic annotation from a message.
func KnowledgeFromMessage(msg protoreflect.MessageDescriptor) (KnowledgeTopic, bool) {
	opts, ok := msg.Options().(*descriptorpb.MessageOptions)
	if !ok || opts == nil {
		return KnowledgeTopic{}, false
	}
	raw := opts.ProtoReflect().GetUnknown()
	ext := findExtension(raw, messageKnowledgeField)
	if ext == nil {
		return KnowledgeTopic{}, false
	}
	return parseKnowledgeTopic(ext), true
}

// ContextFromField extracts the FieldContext annotation from a message field.
func ContextFromField(field protoreflect.FieldDescriptor) (FieldContext, bool) {
	opts, ok := field.Options().(*descriptorpb.FieldOptions)
	if !ok || opts == nil {
		return FieldContext{}, false
	}
	raw := opts.ProtoReflect().GetUnknown()
	ext := findExtension(raw, fieldContextField)
	if ext == nil {
		return FieldContext{}, false
	}
	return parseFieldContext(ext), true
}

// --- Wire format parsing ---

func findExtension(unknown []byte, fieldNumber protowire.Number) []byte {
	for len(unknown) > 0 {
		num, typ, n := protowire.ConsumeTag(unknown)
		if n < 0 {
			return nil
		}
		unknown = unknown[n:]

		switch typ {
		case protowire.VarintType:
			_, m := protowire.ConsumeVarint(unknown)
			if m < 0 {
				return nil
			}
			unknown = unknown[m:]
		case protowire.Fixed32Type:
			_, m := protowire.ConsumeFixed32(unknown)
			if m < 0 {
				return nil
			}
			unknown = unknown[m:]
		case protowire.Fixed64Type:
			_, m := protowire.ConsumeFixed64(unknown)
			if m < 0 {
				return nil
			}
			unknown = unknown[m:]
		case protowire.BytesType:
			b, m := protowire.ConsumeBytes(unknown)
			if m < 0 {
				return nil
			}
			if num == fieldNumber {
				return b
			}
			unknown = unknown[m:]
		default:
			return nil
		}
	}
	return nil
}

func parseKnowledgeTopic(raw []byte) KnowledgeTopic {
	var out KnowledgeTopic
	for len(raw) > 0 {
		num, typ, n := protowire.ConsumeTag(raw)
		if n < 0 {
			return out
		}
		raw = raw[n:]
		if typ == protowire.BytesType {
			b, m := protowire.ConsumeBytes(raw)
			if m < 0 {
				return out
			}
			switch num {
			case 1:
				out.Topic = string(b)
			case 2:
				out.Title = string(b)
			case 3:
				out.Summary = string(b)
			case 4:
				out.Example = string(b)
			}
			raw = raw[m:]
		} else {
			skip := consumeField(typ, raw)
			if skip < 0 {
				return out
			}
			raw = raw[skip:]
		}
	}
	return out
}

func parseEnumValueContext(raw []byte) EnumValueContext {
	var out EnumValueContext
	for len(raw) > 0 {
		num, typ, n := protowire.ConsumeTag(raw)
		if n < 0 {
			return out
		}
		raw = raw[n:]
		if typ == protowire.BytesType {
			b, m := protowire.ConsumeBytes(raw)
			if m < 0 {
				return out
			}
			switch num {
			case 1:
				out.HumanName = string(b)
			case 2:
				out.CommonQueries = string(b)
			case 3:
				out.Topic = string(b)
			}
			raw = raw[m:]
		} else {
			skip := consumeField(typ, raw)
			if skip < 0 {
				return out
			}
			raw = raw[skip:]
		}
	}
	return out
}

func parseFieldContext(raw []byte) FieldContext {
	var out FieldContext
	for len(raw) > 0 {
		num, typ, n := protowire.ConsumeTag(raw)
		if n < 0 {
			return out
		}
		raw = raw[n:]
		if typ == protowire.BytesType {
			b, m := protowire.ConsumeBytes(raw)
			if m < 0 {
				return out
			}
			switch num {
			case 1:
				out.Description = string(b)
			case 2:
				out.Example = string(b)
			}
			raw = raw[m:]
		} else {
			skip := consumeField(typ, raw)
			if skip < 0 {
				return out
			}
			raw = raw[skip:]
		}
	}
	return out
}

func consumeField(typ protowire.Type, raw []byte) int {
	switch typ {
	case protowire.VarintType:
		_, m := protowire.ConsumeVarint(raw)
		return m
	case protowire.Fixed32Type:
		_, m := protowire.ConsumeFixed32(raw)
		return m
	case protowire.Fixed64Type:
		_, m := protowire.ConsumeFixed64(raw)
		return m
	case protowire.BytesType:
		_, m := protowire.ConsumeBytes(raw)
		return m
	default:
		return -1
	}
}
