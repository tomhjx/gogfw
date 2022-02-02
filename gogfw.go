package gogfw

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
	"unicode"
)

type Handle struct {
	data []byte
}

const ITEM_TYPE_IP = "ip"
const ITEM_TYPE_DOMAIN = "domain"
const ITEM_TYPE_DOMAIN_SUFFIX = "domain_suffix"
const ITEM_TYPE_DOMAIN_KEYWORD = "domain_keyword"

type Item struct {
	Type  string
	Value string
}

func OpenOffline(file string) (handle *Handle, err error) {
	d, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return &Handle{data: d}, nil
}

func OpenOnline(url string) (handle *Handle, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("online request returned code: %d", resp.StatusCode)
	}

	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &Handle{data: d}, nil
}

func (i *Handle) ParseItemKeyword(t string) (val string, err error) {
	if !strings.Contains(t, "*") {
		return "", nil
	}
	t = strings.ReplaceAll(t, ".", "*")
	keywords := strings.Split(t, "*")
	maxlen := 0
	for _, v := range keywords {
		if len(v) <= maxlen {
			continue
		}
		val = v
		maxlen = len(val)
	}
	return val, nil
}

func (i *Handle) ParseItem(t string) (item *Item, err error) {
	line := strings.TrimSpace(t)
	if len(t) == 0 {
		return nil, errors.New("ignore empty line")
	}
	if !regexp.MustCompile(`\w+`).MatchString(line) {
		return nil, fmt.Errorf("not enough character, ignore line: %s", line)
	}

	itype := ITEM_TYPE_DOMAIN_KEYWORD
	// log.Printf("source :%s", line)

	// header
	start := 1
	headerc := line[0]
	if unicode.IsNumber(rune(headerc)) || unicode.IsLetter(rune(headerc)) {
		start = 0
	} else {
		switch headerc {
		case '@', '!', '~':
			return item, fmt.Errorf("ignore line: %s", line)
		case '.':
		case '|':
			if line[1] == '|' {
				start = 2
				itype = ITEM_TYPE_DOMAIN_SUFFIX
			}
		default:
			return item, fmt.Errorf("unsupported line: %s", line)
		}
	}

	// http://abc.def.gh.jkl.com/def/gh/jkl => abc.def.gh.jkl.com
	liner := regexp.MustCompile(`/+`)
	linel := liner.Split(line[start:], 3)

	// log.Printf("regexp split:%v", linel)
	linei := 0

	// http:, abc.com, def, gh => abc.com
	// abc.com, def, gh => abc.com
	if len(linel) > 1 && strings.Contains(linel[0], ":") {
		linei = 1
	}
	val := strings.SplitN(linel[linei], ":", 2)[0]

	// log.Printf("step#1.value: %s", val)

	if net.ParseIP(val) != nil {
		return &Item{
			Type:  ITEM_TYPE_IP,
			Value: val,
		}, nil
	}

	// abc.def.gh.jkl.com => jkl.com (domain_suffix)
	linevl := strings.SplitAfter(val, ".")
	linevlen := len(linevl)
	val = linevl[0]
	if linevlen > 1 {
		itype = ITEM_TYPE_DOMAIN_SUFFIX
		val = strings.Join(linevl[linevlen-2:], "")
	}

	// abc.def*gh*jklabc.com => jklabc (domain_keyword)
	keyword, err := i.ParseItemKeyword(val)
	if err != nil {
		return nil, err
	}

	if keyword != "" {
		return &Item{
			Type:  ITEM_TYPE_DOMAIN_KEYWORD,
			Value: keyword,
		}, nil
	}

	// log.Printf("step#2.type: %s", itype)
	// log.Printf("step#2.value: %s", val)

	item = &Item{
		Type:  itype,
		Value: val,
	}
	return item, nil
}

func (i *Handle) ReadItems() (items []*Item, err error) {
	dst := make([]byte, base64.StdEncoding.DecodedLen(len(i.data)))
	n, err := base64.StdEncoding.Decode(dst, i.data)
	if err != nil {
		return nil, err
	}
	data := dst[:n]
	if !bytes.HasPrefix(data, []byte("[AutoProxy ")) {
		return nil, errors.New("invalid auto proxy file")
	}
	exists := make(map[string]bool)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		item, err := i.ParseItem(scanner.Text())
		if err != nil {
			log.Println(err)
			continue
		}
		if item == nil {
			continue
		}
		if exists[item.Value] {
			continue
		}
		exists[item.Value] = true
		items = append(items, item)
	}
	return items, nil
}
