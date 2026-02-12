package transform

import (
	"bytes"
	"io"
	"regexp"

	"golang.org/x/text/transform"
)

type ReplaceFunc func(match []byte) []byte

type regexTransformer struct {
	re      *regexp.Regexp
	replace ReplaceFunc
	buf     bytes.Buffer
	atEOF   bool
}

func Regex(re *regexp.Regexp, replace ReplaceFunc) transform.Transformer {
	return &regexTransformer{
		re:      re,
		replace: replace,
	}
}

func RegexString(pattern string, replace ReplaceFunc) (transform.Transformer, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return Regex(re, replace), nil
}

func RegexLiteral(re *regexp.Regexp, replacement []byte) transform.Transformer {
	return Regex(re, func([]byte) []byte {
		return replacement
	})
}

func RegexStringLiteral(pattern string, replacement string) (transform.Transformer, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return RegexLiteral(re, []byte(replacement)), nil
}

func (t *regexTransformer) Reset() {
	t.buf.Reset()
	t.atEOF = false
}

func (t *regexTransformer) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	t.atEOF = atEOF
	t.buf.Write(src)
	nSrc = len(src)

	data := t.buf.Bytes()

	if !atEOF {
		loc := t.re.FindIndex(data)
		if loc == nil {
			safe := len(data)
			if safe > 0 {
				safe = safe - t.re.NumSubexp() - 1
				if safe < 0 {
					safe = 0
				}
			}
			if safe > 0 && safe <= len(dst) {
				copy(dst, data[:safe])
				t.buf.Reset()
				t.buf.Write(data[safe:])
				return safe, nSrc, nil
			}
			return 0, nSrc, nil
		}
	}

	result := t.re.ReplaceAllFunc(data, t.replace)

	if len(result) > len(dst) {
		copy(dst, result[:len(dst)])
		t.buf.Reset()
		t.buf.Write(result[len(dst):])
		return len(dst), nSrc, transform.ErrShortDst
	}

	copy(dst, result)
	t.buf.Reset()
	return len(result), nSrc, nil
}

type Chain struct {
	transformers []transform.Transformer
}

func NewChain(transformers ...transform.Transformer) *Chain {
	return &Chain{transformers: transformers}
}

func (c *Chain) Add(t transform.Transformer) *Chain {
	c.transformers = append(c.transformers, t)
	return c
}

func (c *Chain) Transformer() transform.Transformer {
	if len(c.transformers) == 0 {
		return transform.Nop
	}
	if len(c.transformers) == 1 {
		return c.transformers[0]
	}
	return transform.Chain(c.transformers...)
}

func (c *Chain) Reader(r io.Reader) io.Reader {
	return transform.NewReader(r, c.Transformer())
}

func (c *Chain) Writer(w io.Writer) io.WriteCloser {
	return transform.NewWriter(w, c.Transformer())
}

func (c *Chain) String(s string) (string, error) {
	result, _, err := transform.String(c.Transformer(), s)
	return result, err
}

func (c *Chain) Bytes(b []byte) ([]byte, error) {
	result, _, err := transform.Bytes(c.Transformer(), b)
	return result, err
}

func String(t transform.Transformer, s string) (string, error) {
	result, _, err := transform.String(t, s)
	return result, err
}

func Bytes(t transform.Transformer, b []byte) ([]byte, error) {
	result, _, err := transform.Bytes(t, b)
	return result, err
}

func Reader(r io.Reader, t transform.Transformer) io.Reader {
	return transform.NewReader(r, t)
}

func Writer(w io.Writer, t transform.Transformer) io.WriteCloser {
	return transform.NewWriter(w, t)
}
