package parsers

import (
	"context"
	"golang.org/x/text/unicode/norm"
	"strconv"
	"strings"

	"github.com/goccha/yubinbango/pkg/domains"
	"github.com/goccha/yubinbango/pkg/entities"

	"github.com/goccha/logging/log"
	"golang.org/x/text/width"
)

type Parser interface {
	Parse(ctx context.Context, rec []string) entities.Yubinbango
}

func NewParser() Parser {
	return &CsvParser{}
}

type CsvParser struct {
}

func (p *CsvParser) Parse(ctx context.Context, row []string) entities.Yubinbango {
	switch len(row) {
	case 13:
		return parseOffice(row)
	default:
		return parse(ctx, row)
	}
}

func parseRange(head, tail, suffix, suffixKana string) ([]string, []string, error) {
	head = width.Fold.String(head)
	tail = width.Fold.String(tail)
	head = strings.TrimSuffix(head, suffix)
	prefix := ""
	top, err := strconv.Atoi(head)
	if err != nil {
		runes := []rune(head)
		for i := len(runes) - 1; i >= 0; i-- {
			if !('0' <= runes[i] && runes[i] <= '9') { // 数字以外を見つけた場合
				num := string(runes[i+1:]) // 末尾の数字を取得
				if top, err = strconv.Atoi(num); err != nil {
					return nil, nil, err
				}
				prefix = width.Widen.String(string(runes[:i+1]))
				break
			}
		}
		if err != nil {
			return nil, nil, err
		}
	}
	var bottom int
	tail = strings.TrimSuffix(tail, suffix)
	var addrs []string
	var addrKana []string
	bottom, err = strconv.Atoi(tail)
	if err != nil {
		addrs = append(addrs, prefix+width.Widen.String(strconv.Itoa(top))+suffix)
		addrs = append(addrs, width.Widen.String(tail)+suffix)

		addrKana = append(addrKana, prefix+width.Widen.String(strconv.Itoa(top))+suffixKana)
		addrKana = append(addrKana, width.Widen.String(tail)+suffixKana)
	} else {
		for i := top; i <= bottom; i++ {
			addrs = append(addrs, prefix+width.Widen.String(strconv.Itoa(i))+suffix)
			addrKana = append(addrKana, prefix+width.Widen.String(strconv.Itoa(i))+suffixKana)
		}
	}
	return addrs, addrKana, nil
}

func splitValue(value, sep string, suffix, suffixKana string) ([]string, []string, error) {
	values := strings.Split(value, sep)
	addrs := make([]string, 0, len(values))
	addrKana := make([]string, 0, len(values))
	for _, v := range values {
		if strings.Contains(v, "〜") {
			rng := strings.Split(v, "〜")
			if list, kana, err := parseRange(rng[0], rng[1], suffix, suffixKana); err != nil {
				return values, nil, err
			} else {
				addrs = append(addrs, list...)
				addrKana = append(addrKana, kana...)
			}
		} else {
			if strings.HasSuffix(v, suffix) {
				addrs = append(addrs, v)
			} else {
				addrs = append(addrs, v+suffix)
			}
			if suffixKana != "" {
				v = strings.TrimSuffix(v, suffix)
				if strings.HasSuffix(v, suffixKana) {
					addrKana = append(addrKana, v)
				} else {
					addrKana = append(addrKana, v+suffixKana)
				}
			}
		}
	}
	return addrs, addrKana, nil
}

func parseExp(exp string) ([]string, []string, error) {
	if !strings.Contains(exp, "「") && !strings.Contains(exp, "＜") && !strings.Contains(exp, "（") { // 「」,＜＞,（）が含まれていない場合
		if strings.Contains(exp, "、") {
			if strings.HasSuffix(exp, "丁目") {
				return splitValue(exp, "、", "丁目", "チョウメ")
			} else if strings.HasSuffix(exp, "チョウメ") {
				return splitValue(exp, "、", "チョウメ", "")
			} else if strings.HasSuffix(exp, "番地") {
				return splitValue(exp, "、", "番地", "バンチ")
			} else if strings.HasSuffix(exp, "バンチ") {
				return splitValue(exp, "、", "バンチ", "")
			} else {
				values := strings.Split(exp, "、")
				addrs := make([]string, len(values))
				copy(addrs, values)
				return addrs, nil, nil
			}
		} else {
			if strings.Contains(exp, "〜") {
				if strings.HasSuffix(exp, "丁目") {
					values := strings.Split(exp, "〜")
					if len(values) == 2 {
						return parseRange(values[0], values[1], "丁目", "チョウメ")
					}
				} else if strings.HasSuffix(exp, "番地") {
					values := strings.Split(exp, "〜")
					if len(values) == 2 {
						return parseRange(values[0], values[1], "番地", "バンチ")
					}
				}
			}
		}
	}
	return []string{exp}, nil, nil
}

func parse(ctx context.Context, row []string) entities.Yubinbango {
	town := row[8]
	townKana := row[5]
	var street []string
	var streetKana []string
	if town == "以下に掲載がない場合" {
		town = ""
		townKana = ""
	} else if index := strings.Index(town, "（"); index > 0 {
		if last := strings.LastIndex(town, "）"); last > 0 {
			exp := town[index+len("（") : last]
			if addrs, kana, err := parseExp(exp); err == nil {
				street = addrs
				streetKana = kana
			}
			town = town[:index]
		}
		if index = strings.Index(townKana, "（"); index > 0 {
			if last := strings.LastIndex(townKana, "）"); last > 0 {
				if len(streetKana) == 0 {
					exp := townKana[index+len("（") : last]
					if addrs, _, err := parseExp(exp); err == nil {
						streetKana = addrs
					}
				}
				townKana = townKana[:index]
			}
		}
	}
	pref := domains.Prefecture(row[6])
	var addresses []entities.Address
	if len(street) > 0 {
		addresses = make([]entities.Address, 0, len(street))
		if len(street) != len(streetKana) {
			streetKana = make([]string, len(street))
			cnt := 0
			for i, a := range street {
				if kana, ok := domains.ParseKyoto(ctx, a); ok {
					log.Debug(ctx).Msgf("ok: %s=%s", a, kana)
					streetKana[i] = kana
					cnt++
				} else {
					log.Error(ctx).Msgf("ng: %s", a)
				}
			}
			if len(street) != cnt {
				log.Error(ctx).Msgf("kanji=%v  kana=%v\n", street, streetKana)
			}
		}
		for i := range street {
			kana := ""
			if i < len(streetKana) {
				kana = streetKana[i]
			}
			addresses = append(addresses, entities.Address{
				City:       row[7],
				Town:       town,
				Street:     street[i],
				CityKana:   row[4],
				TownKana:   townKana,
				StreetKana: kana,
			})
		}
	} else {
		addresses = []entities.Address{
			{
				City:     row[7],
				Town:     town,
				CityKana: row[4],
				TownKana: townKana,
			},
		}
	}
	return entities.Yubinbango{
		ZipCode:   row[2],
		Pref:      pref,
		PrefKana:  pref.Kana(),
		Addresses: addresses,
	}
}

func parseOffice(row []string) entities.Yubinbango {
	//kana := width.Fold.String(row[1])
	kana := norm.NFKC.String(row[1])
	pref := domains.Prefecture(row[3])
	return entities.Yubinbango{
		Pref:     pref,
		PrefKana: pref.Kana(),
		Addresses: []entities.Address{
			{
				City:       row[4],
				Town:       row[5],
				Address:    row[6],
				OfficeKana: kana,
				OfficeName: row[2],
			},
		},
		ZipCode: row[7],
	}
}
