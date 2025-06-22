package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var re = regexp.MustCompile(`[-+]?[\s]*([\d.,]+)\s*([A-Za-z]{0,3})\s*(.*)`)

type ParsedExpense struct {
	UserID   int64
	Amount   float64
	Currency string
	Note     string
}

func ParseExpense(in string) (ParsedExpense, error) {
	m := re.FindStringSubmatch(strings.TrimSpace(in))
	if len(m) != 4 {
		return ParsedExpense{}, fmt.Errorf("не удалось разобрать строку")
	}
	rawAmt := strings.ReplaceAll(m[1], ",", ".")
	amt, err := strconv.ParseFloat(rawAmt, 64)
	if err != nil {
		return ParsedExpense{}, fmt.Errorf("сумма: %w", err)
	}
	cur := strings.ToUpper(m[2])
	if cur == "" {
		cur = "EUR"
	}
	note := strings.TrimSpace(m[3])
	if note == "" {
		note = "Без категории"
	}
	return ParsedExpense{
		Amount:   amt,
		Currency: cur,
		Note:     note,
	}, nil
}
