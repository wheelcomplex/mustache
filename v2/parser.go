package mustache

import (
	"bufio"
	"errors"
	"io"
	"log"
	"strconv"
	"strings"
)

const (
	T_CONS = iota
	T_Val
	T_Section
	T_Comment
	T_Partial
	T_End
)

type tag struct {
	Value string
	Type  int
	Flag  bool
}

func Parse(r io.Reader) (*Template, error) {
	tpl := &Template{}
	tpl.Tree = make([]Node, 0)
	//var err error

	rd := bufio.NewReaderSize(r, 1024*1024)
	lineNumber := -1
	flag := true
	sections := make([]*SectionNode, 0)
	for flag {
		lineNumber++
		line, err := rd.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			flag = false
		}

		tags, err := parseLine(line, lineNumber)
		if err != nil {
			return nil, err
		}

		// Section Tag only?

		_s_count := 0
		var _tag2 tag
		for _, _tag := range tags {
			switch _tag.Type {
			case T_Section:
				_s_count += 1
				_tag2 = _tag
			case T_CONS:
				if strings.Trim(_tag.Value, " \t\r\n") != "" {
					_s_count = -1
					break
				}
			default:
				_s_count = -1
				break
			}
			if _s_count == -1 {
				break
			}
		}
		if _s_count == 1 {
			log.Println("> Single Section > "+_tag2.Value, tags)
			tags = []tag{_tag2}
		}

		for _, _tag := range tags {
			//log.Printf(">>> %v", _tag)
			_ = log.Ldate

			switch _tag.Type {
			case T_Comment:
				continue
			case T_Section:
				log.Printf(">Section [%v]", _tag.Value)
				sec := &SectionNode{_tag.Value, flag, make([]Node, 0)}
				if len(sections) == 0 {
					tpl.Tree = append(tpl.Tree, sec)
					log.Printf("Tree Len=%v", len(tpl.Tree))
				} else {
					sections[len(sections)-1].Clildren = append(sections[len(sections)-1].Clildren, sec)
				}
				sections = append(sections, sec)
			case T_End:
				if len(sections) == 0 || sections[len(sections)-1].name != _tag.Value {
					//log.Printf(">> %v", sections)
					return nil, errors.New("End TAG  Invaild >>" + _tag.Value)
				}
				log.Printf(">Section End [%v]", _tag.Value)
				sections = sections[:len(sections)-1]
			default:
				var node Node
				switch _tag.Type {
				case T_CONS:
					log.Println("Cons ? --> " + _tag.Value)
					node = &ConstantNode{_tag.Value}
				case T_Val:
					node = &ValNode{_tag.Value, _tag.Flag}
				case T_Partial:
					node = &PartialNode{_tag.Value}
				}
				if len(sections) == 0 {
					tpl.Tree = append(tpl.Tree, node)
				} else {
					sections[len(sections)-1].Clildren = append(sections[len(sections)-1].Clildren, node)
				}
			}

		}
	}
	return tpl, nil
}

func parseLine(line string, lineNumber int) (tags []tag, err error) {
	sz := len(line)
	start := 0
	end := 0
	started := false
	tags = make([]tag, 0)
	for end < sz {
		//log.Println("start=", start, "end=", end, "line=", lineNumber)
		started = false

		//Search Start
		for end < sz {
			//log.Println("Search TAG,  start=", start, "end=", end)
			if line[end] == '{' && (end+3) < sz && line[end+1] == '{' {
				started = true
				if end > start {
					tags = append(tags, tag{line[start:end], T_CONS, false})
				}
				start = end + 2
				end = start + 1
				//log.Println("Tag Start Found, beark")
				break
			} else {
				started = false
				if line[end] == '\\' {
					if (end + 1) < sz {
						end++
					}
				}
				end++
			}
		}

		if !started {
			if end > start {
				tags = append(tags, tag{line[start:end], T_CONS, false})
			}
			break
		}

		//log.Println("Search Tag end", "start=", start, "end=", end)

		tagValue := ""

		for (end + 1) < sz {
			if line[end] != '}' || line[end+1] != '}' {
				//log.Println("Current=" + line[end:end+2])
				end++
				continue
			}
			escape := false
			// {{{ABC}}} --> start=2, end=
			if line[start] == '{' && (end+2) < sz && line[end+2] == '}' {
				tagValue = line[start+1 : end]
				start = end + 3
				end = start
				escape = true
			} else {
				tagValue = line[start:end]
				start = end + 2
				end = start
			}

			//log.Println("Tag init Value=" + tagValue)

			tagValue = strings.TrimRight(tagValue, " \t")
			if tagValue == "" {
				return tags, parseError{lineNumber, "Blank Tag"}
			}

			switch tagValue[0] {
			case '&':
				tags = append(tags, tag{tagValue[1:], T_Val, true})
			case '#':
				tags = append(tags, tag{tagValue[1:], T_Section, false})
			case '^':
				tags = append(tags, tag{tagValue[1:], T_Section, true})
			case '>':
				tags = append(tags, tag{tagValue[1:], T_Partial, false})
			case '!':
				tags = append(tags, tag{tagValue[1:], T_Comment, false})
			case '/':
				tags = append(tags, tag{tagValue[1:], T_End, false})
			default:
				tags = append(tags, tag{tagValue, T_Val, escape})
			}
			break
		}
	}
	return tags, nil
}

type parseError struct {
	LineNumber int
	Message    string
}

func (p parseError) Error() string {
	return "Error at Line" + strconv.Itoa(p.LineNumber) + " " + p.Message
}
