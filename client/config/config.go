package config

type Server struct {
	Mysql  Mysql  `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
	Mssql  Mssql  `mapstructure:"mssql" json:"mssql" yaml:"mssql"`
	System System `mapstructure:"system" json:"system" yaml:"system"`
	Redis  Redis  `mapstructure:"redis" json:"redis" yaml:"redis"`
	Zap    Zap    `mapstructure:"zap" json:"zap" yaml:"zap"`
}
