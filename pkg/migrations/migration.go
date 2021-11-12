package migrations

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/thestephenstanton/grules"
)

type v1Composites struct {
	Composites []composite `json:"composites"`
}

type composite struct {
	Operator string `json:"operator"`
	Rules    []rule `json:"rules"`
}

type rule struct {
	Comparator string      `json:"comparator"`
	Path       string      `json:"path"`
	Value      interface{} `json:"value"`
}

// MigrateGrulesV1ToV2 takes the v1CompositesJSON json and turns them to V2
func MigrateGrulesV1ToV2(v1CompositesJSON string) (string, error) {
	var v1Composites v1Composites
	err := json.NewDecoder(strings.NewReader(v1CompositesJSON)).Decode(&v1Composites)
	if err != nil {
		return "", fmt.Errorf("unmarshalling v1Composites: %w", err)
	}

	v2, err := v1Composites.toV2()
	if err != nil {
		return "", fmt.Errorf("converting to v2: %w", err)
	}

	v2Bytes, err := json.Marshal(v2)
	if err != nil {
		return "", fmt.Errorf("marshalling v2: %w", err)
	}

	return string(v2Bytes), nil
}

func (v1 v1Composites) toV2() (grules.Rule, error) {
	if len(v1.Composites) == 0 {
		return grules.Rule{}, errors.New("no composites found")
	}

	// most simple case
	if len(v1.Composites) == 1 && len(v1.Composites[0].Rules) == 1 {
		oldRule := v1.Composites[0].Rules[0]

		newRule := grules.Rule{
			Comparator: oldRule.Comparator,
			Path:       oldRule.Path,
			Value:      oldRule.Value,
		}

		return newRule, nil
	}

	// if there is only 1 composite and multiple rules, we can take 1 depth off
	if len(v1.Composites) == 1 {
		return v1.Composites[0].toRule(), nil
	}

	var newRule grules.Rule
	newRule.Operator = grules.And
	for _, composite := range v1.Composites {
		newRule.Rules = append(newRule.Rules, composite.toRule())
	}

	return newRule, nil
}

func (composite composite) toRule() grules.Rule {
	if len(composite.Rules) == 1 {
		oldRule := composite.Rules[0]

		newRule := grules.Rule{
			Comparator: oldRule.Comparator,
			Path:       oldRule.Path,
			Value:      oldRule.Value,
		}

		return newRule
	}

	var newRule grules.Rule
	newRule.Operator = grules.Operator(composite.Operator)
	for _, oldRule := range composite.Rules {
		newRule.Rules = append(newRule.Rules, grules.Rule{
			Comparator: oldRule.Comparator,
			Path:       oldRule.Path,
			Value:      oldRule.Value,
		})
	}

	return newRule
}
