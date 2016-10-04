package irule

import (
	"bytes"
	"encoding/json"
	"github.com/pr8kerl/f5er/f5"
	"io/ioutil"
	"log"
	"strings"
)

type LBRule f5.LBRule

func GetIrule(irule string, bigip *f5.Device) string {
	err, rule := bigip.ShowRule(irule)
	if err != nil {
		log.Fatalf("Error getting iRule: %s", err)
	}
	return rule.ApiAnonymous
}

func UpdateIruleFile(irulefile, iruledest string, bigip *f5.Device) *f5.LBRule {
	var rule []byte
	err, lbrule := bigip.ShowRule(iruledest)
	if err != nil {
		log.Fatalf("Error getting irule, does it exist?")
	}
	irule, err := ioutil.ReadFile(irulefile)
	if err != nil {
		log.Fatalf("Error reading file: %s, %s", err, irule)
	}
	lbrule.ApiAnonymous = strings.TrimSpace(string(irule[:]))

	rule, err = JSONMarshal(lbrule, true)
	b := bytes.NewBuffer(rule)
	err, res := bigip.UpdateRuleRaw(iruledest, b)
	if err != nil {
		log.Fatalf("Error updating iRule: %s. Error: %s", iruledest, err)
	}
	return res
}

func JSONMarshal(v interface{}, unescape bool) ([]byte, error) {
	b, err := json.Marshal(v)

	if unescape {
		b = bytes.Replace(b, []byte("\\u003c"), []byte("<"), -1)
		b = bytes.Replace(b, []byte("\\u003e"), []byte(">"), -1)
		b = bytes.Replace(b, []byte("\\u0026"), []byte("&"), -1)
	}
	return b, err
}
