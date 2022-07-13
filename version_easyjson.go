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

func easyjson8e52a332DecodeGithubComTealFinanceGarcon(in *jlexer.Lexer, out *lines) {
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
		case "version":
			if in.IsNull() {
				in.Skip()
				out.Version = nil
			} else {
				in.Delim('[')
				if out.Version == nil {
					if !in.IsDelim(']') {
						out.Version = make([]string, 0, 4)
					} else {
						out.Version = []string{}
					}
				} else {
					out.Version = (out.Version)[:0]
				}
				for !in.IsDelim(']') {
					var v1 string
					v1 = string(in.String())
					out.Version = append(out.Version, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
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
func easyjson8e52a332EncodeGithubComTealFinanceGarcon(out *jwriter.Writer, in lines) {
	out.RawByte('{')
	first := true
	_ = first
	if len(in.Version) != 0 {
		const prefix string = ",\"version\":"
		first = false
		out.RawString(prefix[1:])
		{
			out.RawByte('[')
			for v2, v3 := range in.Version {
				if v2 > 0 {
					out.RawByte(',')
				}
				out.String(string(v3))
			}
			out.RawByte(']')
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v lines) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson8e52a332EncodeGithubComTealFinanceGarcon(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v lines) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson8e52a332EncodeGithubComTealFinanceGarcon(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *lines) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson8e52a332DecodeGithubComTealFinanceGarcon(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *lines) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson8e52a332DecodeGithubComTealFinanceGarcon(l, v)
}