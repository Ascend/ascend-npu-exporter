//  Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.

// Package main
package main

import (
	"encoding/pem"
	"flag"
	"fmt"
	"huawei.com/npu-exporter/hwlog"
	"huawei.com/npu-exporter/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const (
	dirPrefix  = "/etc/mindx-dl/"
	timeFormat = "2006-01-02T15-04-05.000"
	onekilo    = 1000
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
	defaultLogFile   = "/var/log/mindx-dl/cert-importer/cert-importer.log"
	cptMap           = map[string]string{
		"ne": "npu-exporter", "am": "access-manager", "tm": "task-manager", "lm": "license-manager", "la": "license-agent",
	}
)

var hwLogConfig = &hwlog.LogConfig{FileMaxSize: hwlog.DefaultFileMaxSize,
	MaxBackups: hwlog.DefaultMaxBackups,
	MaxAge:     hwlog.DefaultMinSaveAge,
}

func main() {
	flag.Parse()
	if version {
		fmt.Printf("cert-importer version: %s \n", hwlog.BuildVersion)
		os.Exit(0)
	}
	stopCH := make(chan struct{})
	defer close(stopCH)
	initHwLogger(stopCH)
	hwlog.RunLog.Infof("start to import certificate and the program version is %s", hwlog.BuildVersion)
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
		"am (access-manager),tm(task-manager),lm(license-manager),la(license agent)")
	flag.BoolVar(&version, "version", false,
		"If true,query the version of the program (default false)")
	flag.StringVar(&hwLogConfig.LogFileName, "logFile", defaultLogFile, "Log file path")
}

func importCertFiles(certFile, keyFile, caFile, crlFile string) {
	valid(certFile, keyFile, caFile, crlFile)
	importCert(certFile, keyFile)
	importCA(caFile)
	importCRL(crlFile)
	hwlog.RunLog.Info("import certificate successfully")
	hwlog.RunLog.Info("please delete the relevant sensitive files once you decide not to use them again.")
	os.Exit(0)
}

func importCert(certFile, keyFile string) {
	keyBlock, err := utils.DecryptPrivateKeyWithPd(keyFile, nil)
	if err != nil {
		hwlog.RunLog.Fatal(err)
	}
	// start to import the  certificate file
	certBytes, err := utils.ReadBytes(certFile)
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
	if err = ioutil.WriteFile(certStore, certBytes, utils.FileMode); err != nil {
		hwlog.RunLog.Fatal("write certBytes to config failed ")
	}
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
	}
}

func valid(certFile string, keyFile string, caFile string, crlFile string) {
	if certFile == "" && keyFile == "" && caFile == "" && crlFile == "" {
		hwlog.RunLog.Fatal("no new certificate files need to be imported")
	}
	if certFile == "" || keyFile == "" {
		hwlog.RunLog.Fatal("need input certFile and keyFile together")
	}
	if encryptAlgorithm != utils.Aes128gcm && encryptAlgorithm != utils.Aes256gcm {
		hwlog.RunLog.Warn("reset invalid encryptAlgorithm ")
		encryptAlgorithm = utils.Aes256gcm
	}
	cp, ok := cptMap[component]
	if !ok {
		hwlog.RunLog.Fatal("the component is invalid")
	}
	component = cp
	keyStore = dirPrefix + component + "/" + utils.KeyStore
	certStore = dirPrefix + component + "/" + utils.CertStore
	caStore = dirPrefix + component + "/" + utils.CaStore
	crlStore = dirPrefix + component + "/" + utils.CrlStore
	passFile = dirPrefix + component + "/" + utils.PassFile
	passFileBackUp = dirPrefix + component + "/" + utils.PassFileBackUp
}

func initHwLogger(stopCh <-chan struct{}) {
	if utils.IsExists(hwLogConfig.LogFileName) {
		fi, err := os.Stat(hwLogConfig.LogFileName)
		if err != nil {
			fmt.Println("check log file status failed")
		}
		if fi.Size() > int64(hwLogConfig.FileMaxSize*onekilo*onekilo) {
			newFile := backupName(hwLogConfig.LogFileName)
			if err := os.Rename(hwLogConfig.LogFileName, newFile); err != nil {
				hwlog.RunLog.Fatal("rotate failed")
			}
			err = os.Chmod(newFile, hwlog.BackupLogFileMode)
			if err != nil {
				hwlog.RunLog.Warn("change mode failed")
			}
		}
	}
	if err := hwlog.InitRunLogger(hwLogConfig, stopCh); err != nil {
		fmt.Printf("hwlog init failed, error is %v", err)
		os.Exit(-1)
	}

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
