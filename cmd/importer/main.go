//  Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.

// Package main
package main

import (
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"huawei.com/npu-exporter/hwlog"
	"huawei.com/npu-exporter/utils"
)

const (
	dirPrefix  = "/etc/mindx-dl/"
	timeFormat = "2006-01-02T15-04-05.000"
	onekilo    = 1000
	hwMindX    = 9000
)

var (
	certFile         string
	keyFile          string
	caFile           string
	crlFile          string
	encryptAlgorithm int
	version          bool
	component        string
	keyStore         string
	certStore        string
	caStore          string
	crlStore         string
	passFile         string
	passFileBackUp   string
	kubeConfig       string
	kubeConfStore    string
	defaultLogFile   = "/var/log/mindx-dl/cert-importer/cert-importer.log"
	cptMap           = map[string]string{
		"ne": "npu-exporter", "am": "access-manager", "lm": "license-manager", "la": "license-agent",
		"hc": "hccl-controller", "dp": "device-plugin", "nd": "noded",
	}
	notDel bool
)

var hwLogConfig = &hwlog.LogConfig{FileMaxSize: hwlog.DefaultFileMaxSize,
	MaxBackups: hwlog.DefaultMaxBackups,
	MaxAge:     hwlog.DefaultMinSaveAge,
}

func main() {
	flag.Parse()
	if version {
		fmt.Printf("cert-importer version: %s \n", hwlog.BuildVersion)
		return
	}
	stopCH := make(chan struct{})
	defer close(stopCH)
	err := initHwLogger(stopCH)
	if err != nil {
		fmt.Printf("hwlog init failed, error is %#v", err)
		return
	}
	importKubeConfig(kubeConfig)
	importCertFiles(certFile, keyFile, caFile, crlFile)
}

func init() {
	flag.StringVar(&caFile, "caFile", "", "The root certificate file path")
	flag.StringVar(&certFile, "certFile", "", "The certificate file path")
	flag.StringVar(&keyFile, "keyFile", "",
		"The key file path,If both the certificate and key file exist,system will enable https")
	flag.StringVar(&crlFile, "crlFile", "", "The offline CRL file path")
	flag.IntVar(&encryptAlgorithm, "encryptAlgorithm", utils.Aes256gcm,
		"Use 8 for aes128gcm,9 for aes256gcm,not recommended config it in general")
	flag.StringVar(&component, "cpt", "ne", "The component name such as ne (npu-exporter),"+
		"am (access-manager),lm(license-manager),la(license agent),hc(hccl-controller),"+
		"dp(device-plugin),nd(noded)")
	flag.BoolVar(&version, "version", false,
		"If true,query the version of the program (default false)")
	flag.StringVar(&hwLogConfig.LogFileName, "logFile", defaultLogFile, "Log file path")
	flag.StringVar(&kubeConfig, "kubeConfig", "", "The k8s config file path")
	flag.BoolVar(&notDel, "n", false,
		"If true,stop delete the sensitive original file automatically")
}

func importCertFiles(certFile, keyFile, caFile, crlFile string) {
	valid(certFile, keyFile, caFile, crlFile)
	hwlog.RunLog.Infof("start to import certificate and the program version is %s", hwlog.BuildVersion)
	importCert(certFile, keyFile)
	importCA(caFile)
	importCRL(crlFile)
	adjustOwner()
	hwlog.RunLog.Info("import certificate successfully")
	if notDel {
		hwlog.RunLog.Info("please delete the relevant sensitive files once you decide not to use them again.")
		return
	}
	if err := utils.OverridePassWdFile(keyFile, []byte{}, utils.FileMode); err != nil {
		hwlog.RunLog.Warn("security delete key file failed")
	}
	err := os.Remove(keyFile)
	if err != nil {
		hwlog.RunLog.Warn("delete private key file automatically failed,please delete it by yourself")
		return
	}
	hwlog.RunLog.Warn("delete private key file automatically")
	return
}

func importCert(certFile, keyFile string) {
	keyBlock, err := utils.DecryptPrivateKeyWithPd(keyFile, nil)
	if err != nil {
		hwlog.RunLog.Fatal(err)
	}
	// start to import the  certificate file
	certBytes, err := utils.ReadLimitBytes(certFile, utils.Size10M)
	if err != nil {
		hwlog.RunLog.Fatal("read certFile failed")
	}
	// validate certification and private key, if not pass, program will exit
	if _, err = utils.ValidateX509Pair(certBytes, pem.EncodeToMemory(keyBlock)); err != nil {
		hwlog.RunLog.Fatal(err)
	}
	if err = utils.MakeSureDir(keyStore); err != nil {
		hwlog.RunLog.Fatal(err)
	}
	// encrypt private key again with passwd
	encryptedBlock, err := utils.EncryptPrivateKeyAgain(keyBlock, passFile, passFileBackUp, encryptAlgorithm)
	if err = utils.OverridePassWdFile(keyStore, pem.EncodeToMemory(encryptedBlock), utils.FileMode); err != nil {
		hwlog.RunLog.Fatal(err)
	}
	hwlog.RunLog.Info("encrypted key file import successfully")
	if err = ioutil.WriteFile(certStore, certBytes, utils.FileMode); err != nil {
		hwlog.RunLog.Fatal("write certBytes to config failed ")
	}
	hwlog.RunLog.Info("cert file import successfully")
}

func importCA(caFile string) {
	// start to import the ca certificate file
	caBytes, err := utils.CheckCaCert(caFile)
	if err != nil {
		hwlog.RunLog.Fatal(err)
	}
	if len(caBytes) != 0 {
		if err = ioutil.WriteFile(caStore, caBytes, utils.FileMode); err != nil {
			hwlog.RunLog.Fatal("write caBytes to config failed ")
		}
		hwlog.RunLog.Info("ca file import successfully")
	}
}

func importCRL(crlFile string) {
	// start to import the crl file
	crlBytes, err := utils.CheckCRL(crlFile)
	if err != nil {
		hwlog.RunLog.Fatal(err)
	}
	if len(crlBytes) != 0 {
		if err = ioutil.WriteFile(crlStore, crlBytes, utils.FileMode); err != nil {
			hwlog.RunLog.Fatal("write crlBytes to config failed ")
		}
		hwlog.RunLog.Info("crl file import successfully")
	}
}

func valid(certFile string, keyFile string, caFile string, crlFile string) {
	if certFile == "" && keyFile == "" && caFile == "" && crlFile == "" {
		hwlog.RunLog.Fatal("no new certificate files need to be imported")
	}
	if certFile == "" || keyFile == "" {
		hwlog.RunLog.Fatal("need input certFile and keyFile together")
	}
	commonValid()
}

func commonValid() {
	if encryptAlgorithm != utils.Aes128gcm && encryptAlgorithm != utils.Aes256gcm {
		hwlog.RunLog.Warn("reset invalid encryptAlgorithm ")
		encryptAlgorithm = utils.Aes256gcm
	}
	cp, ok := cptMap[component]
	if !ok {
		hwlog.RunLog.Fatal("the component is invalid")
	}
	var paths []string
	keyStore = dirPrefix + cp + "/" + utils.KeyStore
	paths = append(paths, keyStore)
	certStore = dirPrefix + cp + "/" + utils.CertStore
	paths = append(paths, certStore)
	caStore = dirPrefix + cp + "/" + utils.CaStore
	paths = append(paths, caStore)
	crlStore = dirPrefix + cp + "/" + utils.CrlStore
	paths = append(paths, crlStore)
	passFile = dirPrefix + cp + "/" + utils.PassFile
	paths = append(paths, passFile)
	passFileBackUp = dirPrefix + cp + "/" + utils.PassFileBackUp
	paths = append(paths, passFileBackUp)
	kubeConfStore = dirPrefix + cp + "/" + utils.KubeCfgFile
	paths = append(paths, kubeConfStore)
	checkPathIsExist(paths)
}

func checkPathIsExist(paths []string) {
	for _, v := range paths {
		if !utils.IsExists(v) {
			continue
		}
		_, err := utils.CheckPath(v)
		if err != nil {
			hwlog.RunLog.Fatal(err)
		}
	}
}

func initHwLogger(stopCh <-chan struct{}) error {
	if utils.IsExists(hwLogConfig.LogFileName) {
		_, err := utils.CheckPath(hwLogConfig.LogFileName)
		if err != nil {
			return err
		}
		fi, err := os.Stat(hwLogConfig.LogFileName)
		if err != nil {
			return err
		}
		if fi.Size() > int64(hwLogConfig.FileMaxSize*onekilo*onekilo) {
			newFile := backupName(hwLogConfig.LogFileName)
			if err := os.Rename(hwLogConfig.LogFileName, newFile); err != nil {
				return err
			}
			err = os.Chmod(newFile, hwlog.BackupLogFileMode)
			if err != nil {
				return err
			}
		}
	}
	if err := hwlog.InitRunLogger(hwLogConfig, stopCh); err != nil {
		return err
	}
	return nil

}

func backupName(name string) string {
	dir := filepath.Dir(name)
	filename := filepath.Base(name)
	suffix := filepath.Ext(filename)
	prefix := filename[:len(filename)-len(suffix)]
	t := time.Now()
	formattedTime := t.Format(timeFormat)
	return filepath.Join(dir, fmt.Sprintf("%s-%s%s", prefix, formattedTime, suffix))
}

func importKubeConfig(kubeConf string) {
	if kubeConf == "" {
		return
	}
	conf, err := utils.CheckPath(kubeConf)
	if err != nil {
		hwlog.RunLog.Fatal(err)
	}
	if suffix := path.Ext(kubeConf); suffix != ".conf" {
		hwlog.RunLog.Fatal("invalid kubeConfig file")
	}
	btes, err := utils.ReadLimitBytes(conf, utils.Size10M)
	if err != nil {
		hwlog.RunLog.Fatal(err)
	}
	commonValid()
	utils.KmcInit(encryptAlgorithm, "", "")
	encryptedConf, err := utils.Encrypt(0, btes)
	if err != nil {
		hwlog.RunLog.Fatal("encrypt kubeConfig failed")
	}
	if err = utils.MakeSureDir(keyStore); err != nil {
		hwlog.RunLog.Fatal(err)
	}
	hwlog.RunLog.Info("encrypt kubeConfig successfully")
	if err := utils.OverridePassWdFile(kubeConfStore, encryptedConf, utils.FileMode); err != nil {
		hwlog.RunLog.Fatal("write encrypted kubeConfig to file failed")
	}
	hwlog.RunLog.Info("import kubeConfig successfully")
	if certFile == "" || keyFile == "" {
		adjustOwner()
		utils.KmcShutDown()
		if notDel {
			hwlog.RunLog.Info("please delete the relevant sensitive files once you decide not to use them again.")
			return
		}
		if err := utils.OverridePassWdFile(conf, []byte{}, utils.FileMode); err != nil {
			hwlog.RunLog.Warn("security delete config failed")
		}
		err = os.Remove(conf)
		if err != nil {
			hwlog.RunLog.Warn("delete config file automatically failed,please delete it by yourself")
			return
		}
		hwlog.RunLog.Info("delete config file automatically")
		return
	}

}

func adjustOwner() {
	path, err := utils.CheckPath(dirPrefix)
	if err != nil {
		hwlog.RunLog.Fatal("config file directory is not safe")
	}
	if err := chownR(path, hwMindX, hwMindX); err != nil {
		hwlog.RunLog.Warn("change file owner failed, please chown to hwMindX manually")
	}
	hwlog.RunLog.Info("change owner successfully")
}
func chownR(path string, uid, gid int) error {
	return filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
		if err == nil {
			err = os.Lchown(name, uid, gid)
		}
		return err
	})
}
