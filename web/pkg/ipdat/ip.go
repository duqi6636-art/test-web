package ipdat

import (
	"encoding/binary"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

type IpInfo struct {
	prefStart [256]uint32
	prefEnd   [256]uint32
	endArr    []uint32
	data      []byte
}

var IPDat *IpInfo
var once sync.Once

// config/ipv4.dat
func GetObject(ipdatFile string) *IpInfo {

	once.Do(func() {
		IPDat = &IpInfo{}
		var err error
		IPDat, err = LoadFile(ipdatFile)
		if err != nil {
			log.Fatal("the IP Dat loaded failed!")
		}
	})
	return IPDat
}

func LoadFile(file string) (*IpInfo, error) {
	p := IpInfo{}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	for k := 0; k < 256; k++ {
		i := k*8 + 4
		p.prefStart[k] = UnpackInt4byte(data[i], data[i+1], data[i+2], data[i+3])
		p.prefEnd[k] = UnpackInt4byte(data[i+4], data[i+5], data[i+6], data[i+7])
	}
	RecordSize := int(UnpackInt4byte(data[0], data[1], data[2], data[3]))
	p.endArr = make([]uint32, RecordSize)
	p.data = data
	for i := 0; i < RecordSize; i++ {
		j := 2052 + (i * 9)
		endipnum := UnpackInt4byte(data[j], data[1+j], data[2+j], data[3+j])
		p.endArr[i] = endipnum
	}

	return &p, err

}

func (p *IpInfo) getAddr(row uint32) string {
	j := 2052 + (row * 9)
	offset := UnpackInt4byte(p.data[4+j], p.data[5+j], p.data[6+j], p.data[7+j])
	length := uint32(p.data[8+j])

	return string(p.data[offset:int(offset+length)])
}

func (p *IpInfo) Get(ip string) (string, error) {
	ips := strings.Split(ip, ".")
	x, _ := strconv.Atoi(ips[0])
	prefix := uint32(x)
	intIP, err := ipToInt(ip)
	if err != nil {
		return "", err
	}

	low := p.prefStart[prefix]
	high := p.prefEnd[prefix]

	var cur uint32
	if low == high {
		cur = low
	} else {
		cur = p.Search(low, high, intIP)
	}
	if cur == 100000000 {
		return "", errors.New("empty")
	} else {
		return p.getAddr(cur), nil
	}

}

func (p *IpInfo) Search(low uint32, high uint32, k uint32) uint32 {
	var M uint32 = 0
	for low <= high {
		mid := (low + high) / 2
		endipNum := p.endArr[mid]
		if endipNum >= k {
			M = mid
			if mid == 0 {
				break
			}
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	return M
}

func ipToInt(ipstr string) (uint32, error) {
	ip := net.ParseIP(ipstr)
	ip = ip.To4()
	if ip == nil {
		return 0, errors.New("ip error")
	}
	return binary.BigEndian.Uint32(ip), nil
}

func UnpackInt4byte(a, b, c, d byte) uint32 {
	return (uint32(a) & 0xFF) | ((uint32(b) << 8) & 0xFF00) | ((uint32(c) << 16) & 0xFF0000) | ((uint32(d) << 24) & 0xFF000000)
}


type IpInfoStruct struct {
	Continent      string `json:"continent"`       //洲
	Country        string `json:"country"`         //国家
	CountryEnglish string `json:"country_english"` //国家英文
	CountryCode    string `json:"country_code"`    //国家英文简写
	Province       string `json:"province"`        //省份
	City           string `json:"city"`            //城市
	District       string `json:"district"`        //区县
	//AreaCode       string `json:"area_code"`       //区域代码
	Isp            string `json:"isp"`             //运营商
	Ip             string `json:"ip"`
	//Longitude      string `json:"longitude"`       //经度
	//Latitude       string `json:"latitude"`        //纬度
	//LocalTime      string `json:"local_time"`      //本地时间
	//Elevation      string `json:"elevation"`       //海拔
	//WeatherStation string `json:"weather_station"` //气象站
	ZipCode        string `json:"zip_code"`        //邮编
	//Version        string `json:"version"`         //版本
	//IsProxy        string `json:"is_proxy"`        //是否代理
	//ProxyType      string `json:"proxy_type"`      //代理类型
	CityCode       string `json:"city_code"`       //城市代码
	Asn            string `json:"asn"`
	//Domain         string `json:"domain"`     //领域
	//UsageType      string `json:"usage_type"` //使用场景
	//Street         string `json:"street"`     //街道
	//ProxyTime      string `json:"proxy_time"` //使用场景
	Hosting        string `json:"hosting"`    //是否是数据中心
}

func (ip *IpInfoStruct)String() string {
	return ip.Continent + "|" + ip.Country + "|" + ip.Province + "|" + ip.City + "|" + ip.Isp + "|" + ip.ZipCode
}

func (obj *IpInfo) GetIpInfo(ip string) (IpInfoStruct, error) {

	ipinfo := IpInfoStruct{}

	ipInfoStr, err := obj.Get(ip)
	if err != nil {
		return ipinfo, err
	}
	infos := strings.Split(ipInfoStr, "|")
	if len(infos) < 6 {
		return ipinfo, errors.New("empty")
	}
	ipinfo = IpInfoStruct{
		Country:        infos[2],
		Province:       infos[1],
		City:           infos[0],
		Isp:            infos[5],
		CountryCode:    infos[2],
		Ip:             ip,
		ZipCode:        infos[3],
		Asn:            infos[4],
		Hosting:        infos[6],
	}
	return ipinfo ,nil

}