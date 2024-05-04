package entities

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/goccha/yubinbango/pkg/domains"

	"github.com/goccha/envar"
	"github.com/goccha/logging/log"
)

type File struct {
	Key  string
	Ext  string
	List []string
	Map  map[string]*Yubinbango
	dict map[string]string
}

func OpenFile(ctx context.Context, path, name string) (*File, error) {
	names := strings.Split(name, ".")
	f := &File{Key: names[0], Ext: names[1]}
	file, err := f.Read(ctx, path, false)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()
	return f, nil
}

func (f *File) MakeDict() {
	f.dict = f.makeDict()
}

func (f *File) makeDict() map[string]string {
	dict := make(map[string]string)
	for _, v := range f.Map {
		for _, a := range v.Addresses {
			if a.OfficeKana != "" {
				dict[a.OfficeName] = a.OfficeKana
			}
			if a.CityKana != "" {
				dict[a.City] = a.CityKana
			}
			if a.TownKana != "" {
				dict[a.Town] = a.TownKana
			}
			if a.StreetKana != "" {
				dict[a.Street] = a.StreetKana
			}
		}
	}
	return dict
}

func (f *File) Add(ctx context.Context, yb *Yubinbango) {
	if f.Map == nil {
		f.Map = make(map[string]*Yubinbango)
	}
	if f.List == nil {
		f.List = make([]string, 0)
	}
	if !slices.Contains(f.List, yb.ZipCode) {
		f.List = append(f.List, yb.ZipCode)
	}
	if _, ok := f.Map[yb.ZipCode]; ok {
		log.Debug(ctx).Msgf("duplicate key: %v / %v", yb.ZipCode, yb)
		f.Map[yb.ZipCode] = f.Map[yb.ZipCode].Merge(*yb.Replenish(f.dict))
	} else {
		f.Map[yb.ZipCode] = yb.Replenish(f.dict)
	}
	f.MakeDict()
}

func (f *File) Read(ctx context.Context, path string, renew bool) (*os.File, error) {
	fileName := fmt.Sprintf("%s/%s.%s", path, f.Key, f.Ext)
	var file *os.File
	if _, err := os.Stat(fileName); err != nil {
		if os.IsNotExist(err) {
			log.Info(ctx).Msgf("file not found: %s", fileName)
			f.dict = make(map[string]string)
			return os.Create(fileName)
		} else {
			return nil, err
		}
	} else {
		log.Debug(ctx).Msgf("file found: %s", fileName)
		file, err = os.OpenFile(fileName, os.O_RDWR, 0644)
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}

		m := make(map[string]*Yubinbango)
		if err = json.Unmarshal(data, &m); err != nil {
			return nil, err
		}
		for k, v := range f.Map {
			if vv, ok := m[k]; !ok {
				m[k] = v
			} else {
				m[k] = v.Merge(*vv)
			}
		}
		f.List = make([]string, 0, len(m))
		for k := range m {
			f.List = append(f.List, k)
		}
		sort.Slice(f.List, func(i, j int) bool {
			return f.List[i] < f.List[j]
		})
		f.Map = m
		f.dict = f.makeDict()
		if renew {
			_ = file.Close()
			file, err = os.Create(fileName)
			if err != nil {
				return nil, err
			}
		}
	}
	return file, nil
}

func (f *File) Write(ctx context.Context, path string, renew bool) (err error) {
	var file *os.File
	file, err = f.Read(ctx, path, renew)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()
	if data, err := marshalJson(f.Map); err != nil {
		return err
	} else {
		if _, err = file.Write(data); err != nil {
			return err
		}
	}
	return nil
}

func marshalJson(m map[string]*Yubinbango) (data []byte, err error) {
	if indent := envar.String("MARSHAL_JSON_INDENT"); indent != "" {
		if data, err = json.MarshalIndent(m, "", indent); err != nil {
			return nil, err
		}
	} else {
		if data, err = json.Marshal(m); err != nil {
			return nil, err
		}
	}
	return
}

type FileFormatter interface {
	Format(f File) (string, error)
}

type JsonFormat struct{}

func (f *JsonFormat) Format(file File) (string, error) {
	bin, err := json.Marshal(file.Map)
	if err != nil {
		return "", err
	}
	return string(bin), nil
}

type JsMarshaller struct {
	Pref        domains.Prefecture `json:"prefecture,omitempty"`
	City        []string           `json:"city,omitempty"`
	Town        []string           `json:"town,omitempty"`
	Address     []string           `json:"address,omitempty"`
	PrefKana    string             `json:"prefecture_kana,omitempty"`
	CityKana    []string           `json:"city_kana,omitempty"`
	TownKana    []string           `json:"town_kana,omitempty"`
	AddressKana []string           `json:"address_kana,omitempty"`
	OfficeName  []string           `json:"office_name,omitempty"`
	OfficeKana  []string           `json:"office_kana,omitempty"`
}

func (j *JsMarshaller) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{
		j.Pref.Id(),
		deduplication(j.City),
		deduplication(j.Town),
		j.Address,
		deduplication(j.CityKana),
		deduplication(j.TownKana),
		j.AddressKana,
		j.OfficeName,
		j.OfficeKana,
	})
}

func deduplication(values []string) []string {
	m := make(map[string]struct{})
	for _, v := range values {
		m[v] = struct{}{}
	}
	if len(m) == 1 {
		return []string{values[0]}
	}
	return values
}

type JsFormat struct{}

func (f *JsFormat) Format(file *File) (string, error) {
	m := make(map[string]*JsMarshaller)
	for _, k := range file.List {
		yb := file.Map[k]
		if w, ok := m[k]; ok {
			for _, v := range yb.Addresses {
				w.City = append(w.City, v.City)
				w.Town = append(w.Town, v.Town)
				w.Address = append(w.Address, v.Street)
				w.CityKana = append(w.CityKana, v.CityKana)
				w.TownKana = append(w.TownKana, v.TownKana)
				w.AddressKana = append(w.AddressKana, v.StreetKana)
				w.OfficeName = append(w.OfficeName, v.OfficeName)
				w.OfficeKana = append(w.OfficeKana, v.OfficeKana)
			}
		} else {
			w = &JsMarshaller{
				Pref:        yb.Pref,
				PrefKana:    yb.PrefKana,
				City:        make([]string, 0, len(yb.Addresses)),
				Town:        make([]string, 0, len(yb.Addresses)),
				Address:     make([]string, 0, len(yb.Addresses)),
				CityKana:    make([]string, 0, len(yb.Addresses)),
				TownKana:    make([]string, 0, len(yb.Addresses)),
				AddressKana: make([]string, 0, len(yb.Addresses)),
				OfficeName:  make([]string, 0, len(yb.Addresses)),
				OfficeKana:  make([]string, 0, len(yb.Addresses)),
			}
			for _, v := range yb.Addresses {
				w.City = append(w.City, v.City)
				w.Town = append(w.Town, v.Town)
				if v.Address != "" {
					w.Address = append(w.Address, v.Address)
				} else {
					w.Address = append(w.Address, v.Street)
				}
				w.CityKana = append(w.CityKana, v.CityKana)
				w.TownKana = append(w.TownKana, v.TownKana)
				if v.AddressKana != "" {
					w.AddressKana = append(w.AddressKana, v.AddressKana)
				} else {
					w.AddressKana = append(w.AddressKana, v.StreetKana)
				}
				w.OfficeName = append(w.OfficeName, v.OfficeName)
				w.OfficeKana = append(w.OfficeKana, v.OfficeKana)
			}
			m[k] = w
		}

	}
	bin, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)

	if _, err = buf.WriteString("$yubin("); err != nil {
		return "", err
	}
	if _, err = buf.Write(bin); err != nil {
		return "", err
	}
	if _, err = buf.WriteString(");"); err != nil {
		return "", err
	}
	return buf.String(), nil
}

type Yubinbango struct {
	ZipCode   string             `json:"zip_code,omitempty"`
	Pref      domains.Prefecture `json:"prefecture,omitempty"`
	PrefKana  string             `json:"prefecture_kana,omitempty"`
	Addresses []Address          `json:"addresses,omitempty"`
}

func (y *Yubinbango) Merge(yb Yubinbango) *Yubinbango {
	if y.ZipCode[:3] != yb.ZipCode[:3] {
		return y
	}
	if y.Pref != yb.Pref {
		return y
	}
	for _, v := range yb.Addresses {
		if !slices.ContainsFunc(y.Addresses, func(a Address) bool {
			return v.Equal(a)
		}) {
			for _, a := range yb.Addresses {
				if a.TownKana == "" {
					if index := slices.IndexFunc(y.Addresses, func(aa Address) bool {
						return aa.City == a.City && aa.Town == a.Town
					}); index >= 0 {
						a.CityKana = y.Addresses[index].CityKana
						a.TownKana = y.Addresses[index].TownKana
					}
				}
			}
			y.Addresses = append(y.Addresses, v)
		}
	}
	return y
}
func (y *Yubinbango) Replenish(dict map[string]string) *Yubinbango {
	if len(dict) == 0 {
		return y
	}
	for i, a := range y.Addresses {
		if a.CityKana == "" {
			if v, ok := dict[a.City]; ok {
				y.Addresses[i].CityKana = v
			}
		}
		if a.TownKana == "" {
			if v, ok := dict[a.Town]; ok {
				y.Addresses[i].TownKana = v
			}
		}
		if a.StreetKana == "" {
			if v, ok := dict[a.Street]; ok {
				y.Addresses[i].StreetKana = v
			}
		}
		if a.OfficeKana == "" {
			if v, ok := dict[a.OfficeName]; ok {
				y.Addresses[i].OfficeKana = v
			}
		}
	}
	return y
}

type Address struct {
	City        string `json:"city,omitempty"`
	Town        string `json:"town,omitempty"`
	Street      string `json:"street,omitempty"`
	Address     string `json:"address,omitempty"`
	CityKana    string `json:"city_kana,omitempty"`
	TownKana    string `json:"town_kana,omitempty"`
	StreetKana  string `json:"street_kana,omitempty"`
	AddressKana string `json:"address_kana,omitempty"`
	OfficeName  string `json:"office_name,omitempty"`
	OfficeKana  string `json:"office_kana,omitempty"`
}

func (a Address) Equal(b Address) bool {
	return a.City == b.City && a.Town == b.Town && a.Street == b.Street && a.OfficeName == b.OfficeName
}
