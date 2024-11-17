package x

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
	"sync"

	"github.com/Masterminds/sprig/v3"
	"github.com/gobwas/glob"
	"github.com/sirupsen/logrus"
)

var _ Selector = (*Selectors)(nil)

var _ Selector = (*SelectorField)(nil)

// Selector is a interface for value compare
type Selector interface {
	Init() error
	Match(obj interface{}) bool
}

type Selectors []*SelectorField

func (ss Selectors) String() string {
	res, _ := json.Marshal(ss)
	return string(res)
}

func (ss Selectors) Init() error {
	for _, s := range ss {
		if err := s.Init(); err != nil {
			return err
		}
	}
	return nil
}

// Match all items must be true
func (ss Selectors) Match(obj interface{}) bool {
	for _, s := range ss {
		if !s.Match(obj) {
			return false
		}
	}
	return true
}

// SelectorField is a struct for selector
type SelectorField struct {
	Key      string       `json:"key,omitempty" yaml:"key"`
	Operator OperatorType `json:"operator,omitempty" yaml:"operator"`
	Values   []string     `json:"values,omitempty" yaml:"values"`

	tplStr string
	tpl    *template.Template

	sync.Once
}

func (s *SelectorField) Init() error {
	var err error

	if s.Operator == "" {
		s.Operator = OperatorIn
	}

	// validate the operator
	err = s.Operator.Validate(s.Values)

	s.Do(func() {
		s.tplStr = prepareTplPathKey(s.Key)
		s.tpl, err = template.New(s.Key).Funcs(sprig.FuncMap()).Parse(s.tplStr)
		if err != nil {
			logrus.Debugf("create tpl %s error %s", s.tplStr, err)
			return
		}

		// try to init all glob pattern
		// if we are glob pattern mode
		for _, v := range s.Values {
			glob.MustCompile(v)
		}
	})

	return err
}

func (s *SelectorField) Match(obj interface{}) bool {
	// get the value out with key
	val, err := getValueByPath(s.tpl, obj)
	if err != nil {
		logrus.Errorf("parse value from tpl %s error %v", s.Key, err)
		return false
	}
	return s.Operator.Match(fmt.Sprintf("%s", val), s.Values)
}

func prepareTplPathKey(s string) string {
	ss := strings.TrimSpace(s)

	// check if we are a full template
	if len(s) > 4 && s[0] == '{' && s[1] == '{' && s[len(s)-1] == '}' && s[len(s)-2] == '}' {
		return ss
	}

	// add dot for value picker
	if len(s) > 0 && s[0] != '.' {
		ss = "." + s
	}
	return fmt.Sprintf("{{ %s }}", ss)
}

func getValueByPath(tpl *template.Template, obj interface{}) (interface{}, error) {
	defer func() {
		if err := recover(); err != nil {
			logrus.Errorf("parse value from tpl %s error %v", tpl.Name(), err)
		}
	}()

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, obj); err != nil {
		return nil, err
	}

	return buf.String(), nil
}

func GetValueByPath(obj interface{}, key string) (interface{}, error) {
	tpl, err := template.New(key).Funcs(sprig.FuncMap()).Parse(prepareTplPathKey(key))
	if err != nil {
		return nil, err
	}

	return getValueByPath(tpl, obj)
}
