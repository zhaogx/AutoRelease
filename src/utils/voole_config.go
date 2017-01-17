package utils

import (
	"encoding/json"
	"fmt"
	"os"
	_ "strconv"
)

type VooleConfigMgmt struct {
	name    string                 "config file name"
	pf      *os.File               "file ptr"
	decoder *json.Decoder          "json decoder ptr"
	content map[string]interface{} "json file content"
}

type VooleConfigCb func(v map[string]interface{}) int

func VooleConfigInit(name string) *VooleConfigMgmt {
	mgmt := new(VooleConfigMgmt)
	mgmt.name = name
	pf, err := os.Open(mgmt.name)
	if err != nil {
		return nil
	}
	mgmt.pf = pf

	mgmt.decoder = json.NewDecoder(mgmt.pf)
	err = mgmt.decoder.Decode(&mgmt.content)
	if err != nil {
		goto exit
	}
	fmt.Println(mgmt.content)
	return mgmt
exit:
	mgmt.pf.Close()
	return nil
}

func (this *VooleConfigMgmt) Destory() {
	if this.pf != nil {
		this.pf.Close()
	}
}

func (this *VooleConfigMgmt) getValue(k ...string) (interface{}, bool) {
	var result interface{}
	var p map[string]interface{}

	result = this.content
	for _, key := range k {
		switch v := result.(type) {
		case map[string]interface{}:
			p = v
		default:
			goto exit
		}
		value, ok := p[key]
		if ok == false {
			goto exit
		}
		result = value
	}
	return result, true
exit:
	return nil, false
}

func (this *VooleConfigMgmt) getNumber(k ...string) (float64, bool) {
	result, f := this.getValue(k...)
	if f == false {
		return 0, false
	}
	switch v := result.(type) {
	case float64:
		return v, true
	}
	return 0, false
}

func (this *VooleConfigMgmt) GetInt(k ...string) (int, bool) {
	result, f := this.getNumber(k...)
	return int(result), f
}
func (this *VooleConfigMgmt) GetInt8(k ...string) (int8, bool) {
	result, f := this.getNumber(k...)
	return int8(result), f
}
func (this *VooleConfigMgmt) GetInt16(k ...string) (int16, bool) {
	result, f := this.getNumber(k...)
	return int16(result), f
}
func (this *VooleConfigMgmt) GetInt32(k ...string) (int32, bool) {
	result, f := this.getNumber(k...)
	return int32(result), f
}
func (this *VooleConfigMgmt) GetInt64(k ...string) (int64, bool) {
	result, f := this.getNumber(k...)
	return int64(result), f
}

func (this *VooleConfigMgmt) GetUint(k ...string) (uint, bool) {
	result, f := this.getNumber(k...)
	return uint(result), f
}
func (this *VooleConfigMgmt) GetUint8(k ...string) (uint8, bool) {
	result, f := this.getNumber(k...)
	return uint8(result), f
}
func (this *VooleConfigMgmt) GetUint16(k ...string) (uint16, bool) {
	result, f := this.getNumber(k...)
	return uint16(result), f
}
func (this *VooleConfigMgmt) GetUint32(k ...string) (uint32, bool) {
	result, f := this.getNumber(k...)
	return uint32(result), f
}
func (this *VooleConfigMgmt) GetUint64(k ...string) (uint64, bool) {
	result, f := this.getNumber(k...)
	return uint64(result), f
}
func (this *VooleConfigMgmt) GetFloat32(k ...string) (float32, bool) {
	result, f := this.getNumber(k...)
	return float32(result), f
}
func (this *VooleConfigMgmt) GetFloat64(k ...string) (float64, bool) {
	return this.getNumber(k...)
}

func (this *VooleConfigMgmt) GetString(k ...string) (string, bool) {
	result, f := this.getValue(k...)
	if f == false {
		goto exit
	}
	switch v := result.(type) {
	case string:
		return v, true
	}
exit:
	return "", false
}

func ReadConfig(name string, v interface{}) error {
	r, err := os.Open(name)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(r)
	err = decoder.Decode(v)
	if err != nil {
		return err
	}
	return nil
}
