package protoparser

import (
	"fmt"
	"text/scanner"
)

// comment\npackage...
// comment\nservice...
// comment\nmessage...
func parse(lex *lexer) (*ProtocolBuffer, error) {
	var pkg string
	service := &Service{}
	var messages []*Message
	for lex.token != scanner.EOF {
		comments := parseComments(lex)

		switch lex.text() {
		case "package":
			p, err := parsePackage(lex)
			if err != nil {
				return nil, err
			}
			pkg = p
		case "service":
			s, err := parseService(lex)
			if err != nil {
				return nil, err
			}
			s.Comments = append(s.Comments, comments...)
			service = s
		case "message":
			message, err := parseMessage(lex)
			if err != nil {
				return nil, err
			}
			message.Comments = append(message.Comments, comments...)
			messages = append(messages, message)
		default:
			lex.next()
			continue
		}
	}
	return &ProtocolBuffer{
		Package:  pkg,
		Service:  service,
		Messages: messages,
	}, nil
}

// "message" var '{' messageContent '}'
func parseMessage(lex *lexer) (*Message, error) {
	text := lex.text()
	if text != "message" {
		return nil, fmt.Errorf("not found message, text=%s", text)
	}

	// メッセージ名を取得する {
	lex.next()
	name := lex.text()
	lex.next()
	// }

	// メッセージの中身を取得する {
	/// '{' を消費する {
	lex.next()
	/// }
	fields, nests, enums, oneofs, err := parseMessageContent(lex)
	if err != nil {
		return nil, err
	}
	// }

	// '}' を消費する {
	lex.next()
	// }

	return &Message{
		Name:   name,
		Fields: fields,
		Nests:  nests,
		Enums:  enums,
		Oneofs: oneofs,
	}, nil
}

// "message"
// "enum"
// field
func parseMessageContent(lex *lexer) (fields []*Field, messages []*Message, enums []*Enum, oneofs []*Oneof, err error) {
	for lex.text() != "}" {
		if lex.token != scanner.Comment {
			return nil, nil, nil, nil, fmt.Errorf("not found comment, text=%s", lex.text())
		}
		comments := parseComments(lex)

		switch lex.text() {
		case "message":
			message, parseErr := parseMessage(lex)
			if parseErr != nil {
				return nil, nil, nil, nil, parseErr
			}
			message.Comments = append(message.Comments, comments...)
			messages = append(messages, message)
		case "enum":
			enum, parseErr := parseEnum(lex)
			if parseErr != nil {
				return nil, nil, nil, nil, parseErr
			}
			enum.Comments = append(enum.Comments, comments...)
			enums = append(enums, enum)
		case "oneof":
			oneof, parseErr := parseOneof(lex)
			if parseErr != nil {
				return nil, nil, nil, nil, parseErr
			}
			oneof.Comments = append(oneof.Comments, comments...)
			oneofs = append(oneofs, oneof)
		default:
			field := parseField(lex)
			field.Comments = append(field.Comments, comments...)
			fields = append(fields, field)
		}
	}

	return fields, messages, enums, oneofs, nil
}
