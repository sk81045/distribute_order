package config

type System struct {
	Name              string `mapstructure:"name" json:"name" yaml:"name"`
	Version           string `mapstructure:"version" json:"version" yaml:"version"`
	Author            string `mapstructure:"author" json:"author" yaml:"author"`
	UpdateLog         string `mapstructure:"update-log" json:"update-log" yaml:"update-log"`
	Env               string `mapstructure:"env" json:"env" yaml:"env"`                                 // 环境值
	Addr              int    `mapstructure:"addr" json:"addr" yaml:"addr"`                              // 端口值
	DbType            string `mapstructure:"db-type" json:"dbType" yaml:"db-type"`                      // 数据库类型:mysql(默认)|sqlite|sqlserver|postgresql
	OssType           string `mapstructure:"oss-type" json:"ossType" yaml:"oss-type"`                   // Oss类型
	UseMultipoint     bool   `mapstructure:"use-multipoint" json:"useMultipoint" yaml:"use-multipoint"` // 多点登录拦截
	LimitCountIP      int    `mapstructure:"iplimit-count" json:"iplimitCount" yaml:"iplimit-count"`
	LimitTimeIP       int    `mapstructure:"iplimit-time" json:"iplimitTime" yaml:"iplimit-time"`
	School            string `mapstructure:"school" json:"school" yaml:"school"`
	SchoolId          string `mapstructure:"school-id" json:"school-id" yaml:"school-id"`
	SellfoodApiType   string `mapstructure:"sellfood-api-type" json:"sellfood-api-type" yaml:"sellfood-api-type"`
	SellfoodApiurl    string `mapstructure:"sellfood-apiurl" json:"sellfood-apiurl" yaml:"sellfood-apiurl"`
	SellfoodAppid     string `mapstructure:"sellfood-appid" json:"sellfood-appid" yaml:"sellfood-appid"`
	SellfoodSecretkey string `mapstructure:"sellfood-secretkey" json:"sellfood-secretkey" yaml:"sellfood-secretkey"`
	SellfoodSoap      string `mapstructure:"sellfood-soap" json:"sellfood-soap" yaml:"sellfood-soap"`
}
