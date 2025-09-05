package internal

import "sync"

var cacheDsn map[string]string = map[string]string{}

func setDsnInternal(key, value string) {
	cacheDsn[key] = value
}
func GetDsn(key string) string {
	return cacheDsn[key]
}

type initSaveDns struct {
	once sync.Once
}

var cacheSaveDns sync.Map

func SetDsn(dnsEncrypt, dns string) {
	actually, _ := cacheSaveDns.LoadOrStore(dnsEncrypt, &initSaveDns{})
	item := actually.(*initSaveDns)
	item.once.Do(func() {
		setDsnInternal(dnsEncrypt, dns)

	})

}
