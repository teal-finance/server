// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package garcon

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjsonE873f711DecodeGithubComTealFinanceGarcon(in *jlexer.Lexer, out *msg) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "message":
			out.Message = string(in.String())
		case "doc":
			out.Doc = string(in.String())
		case "path":
			out.Path = string(in.String())
		case "query":
			out.Query = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonE873f711EncodeGithubComTealFinanceGarcon(out *jwriter.Writer, in msg) {
	out.RawByte('{')
	first := true
	_ = first
	if in.Message != "" {
		const prefix string = ",\"message\":"
		first = false
		out.RawString(prefix[1:])
		out.String(string(in.Message))
	}
	if in.Doc != "" {
		const prefix string = ",\"doc\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Doc))
	}
	if in.Path != "" {
		const prefix string = ",\"path\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Path))
	}
	if in.Query != "" {
		const prefix string = ",\"query\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Query))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v msg) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonE873f711EncodeGithubComTealFinanceGarcon(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v msg) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonE873f711EncodeGithubComTealFinanceGarcon(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *msg) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonE873f711DecodeGithubComTealFinanceGarcon(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *msg) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonE873f711DecodeGithubComTealFinanceGarcon(l, v)
}