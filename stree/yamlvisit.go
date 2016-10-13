package stree

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"regexp"
	"strings"

	log "github.com/cihub/seelog"
)

var yamlLine *regexp.Regexp = regexp.MustCompile(`^(( *)(\-\s+)?([^ ]+)\s*:\s*)(.+)?`)

type yamlKey struct {
	name string
	idx  int
}

func newYamlKey(name string) yamlKey {
	return yamlKey{name: name, idx: -1}
}

func (k *yamlKey) String() string {
	if k.idx < 0 {
		return "." + k.name
	} else {
		return "." + fmt.Sprintf("%s[%d]", k.name, k.idx)
	}
}

func printYamlKeys(keys []yamlKey) (s string) {
	for _, k := range keys {
		s += k.String()
	}
	return
}

func YamlVisit(data io.Reader, visitor func(header, val, key string)) error {

	var indentSize, lineNum int
	var keyList []yamlKey
	var err error

	scanner := bufio.NewScanner(data)
	for scanner.Scan() {
		line := scanner.Text()
		lineNum += 1
		toks := yamlLine.FindStringSubmatch(line)
		if toks == nil {
			continue
		}

		header := toks[1]
		curIndent := len(toks[2])
		isListItem := (len(toks[3]) > 0)
		keyCur := strings.Replace(toks[4], ".", `\.`, -1)
		val := toks[5]
		if indentSize < 1 {
			indentSize = curIndent
		} else if int(math.Mod(float64(curIndent), float64(indentSize))) != 0 {
			return fmt.Errorf("yamlNav unexpected indent size on line %d", lineNum)
		}

		lastIndent := 0
		if len(keyList) > 0 {
			lastIndent = (len(keyList) - 1) * indentSize
		}
		log.Tracef("yamlNav - line: %d  lastIndent: %d  curIndent: %d", lineNum, lastIndent, curIndent)

		if isListItem {
			keyList, err = yamlListItem(keyList, keyCur, lineNum, curIndent, lastIndent)
			if err != nil {
				return err
			}
		} else if curIndent < lastIndent {
			keyList, err = yamlOutdent(keyList, keyCur, lineNum, curIndent, lastIndent, indentSize)
			if err != nil {
				return err
			}
		} else if (curIndent > lastIndent) || (len(keyList) < 1) {
			log.Tracef("yamlNav line %d indent", lineNum)
			keyList = append(keyList, newYamlKey(keyCur))
		} else {
			log.Tracef("yamlNav line %d new key", lineNum)
			keyList[len(keyList)-1].name = keyCur
			keyList[len(keyList)-1].idx = -1
		}

		visitor(header, val, printYamlKeys(keyList))
	}
	if err := scanner.Err(); err != nil {
		return log.Errorf("yamlNav scanner error: %v", err)
	}

	return nil
}

func yamlListItem(keyList []yamlKey, keyCur string, lineNum, curIndent, lastIndent int) ([]yamlKey, error) {
	if curIndent == lastIndent {
		log.Tracef("yamlNav line %d new list", lineNum)
		if len(keyList) < 1 {
			return keyList, fmt.Errorf("yamlNav unlabel list - line: %d", lineNum)
		}
		keyList[len(keyList)-1].idx += 1
		keyList = append(keyList, newYamlKey(keyCur))
	} else {
		log.Tracef("yamlNav line %d continuing list", lineNum)
		if len(keyList) < 2 {
			return keyList, fmt.Errorf("yamlNav list missing parent - line: %d", lineNum)
		}
		keyList[len(keyList)-2].idx += 1
		keyList[len(keyList)-1].name = keyCur
	}
	return keyList, nil
}

func yamlOutdent(keyList []yamlKey, keyCur string, lineNum, curIndent, lastIndent, indentSize int) ([]yamlKey, error) {
	log.Tracef("yamlNav line %d outdent - curIndent: %d  lastIndent: %d  indentSize: %d", lineNum, curIndent, lastIndent, indentSize)
	if len(keyList) < 1 {
		return keyList, fmt.Errorf("yamlNav unexpected outdent - line: %d", lineNum)
	}
	for curIndent < lastIndent {
		keyList = keyList[:len(keyList)-1]
		lastIndent -= indentSize
	}
	if len(keyList) > 0 {
		keyList[len(keyList)-1].name = keyCur
		keyList[len(keyList)-1].idx = -1
	}
	return keyList, nil
}
