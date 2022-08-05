//  Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.

// Package main
package main

import (
	"context"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"syscall"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/npu-exporter/utils"
	"huawei.com/npu-exporter/versions"
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
		fmt.Printf("[OP]cert-importer version: %s \n", versions.BuildVersion)
		return
	}
	err := initHwLogger()
	if err != nil {
		fmt.Printf("hwlog init failed, error is %#v", err)
		return
	}
	name, err := os.Hostname()
	if err != nil {
		hwlog.RunLog.Warn("get hostName failed")
	}
	hwlog.RunLog.Infof("[OP]current userID is %d,hostName is %s,127.0.0.1", syscall.Getuid(), name)
	if err = importKubeConfig(kubeConfig); err != nil {
		hwlog.RunLog.Error(err)
		return
	}
	if err = importCertFiles(certFile, keyFile, caFile, crlFile); err != nil {
		hwlog.RunLog.Error(err)
	}
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

func importCertFiles(certFile, keyFile, caFile, crlFile string) error {
	if err := valid(certFile, keyFile, caFile, crlFile); err != nil {
		return err
	}
	hwlog.RunLog.Infof("[OP]start to import certificate and the program version is %s", versions.BuildVersion)
	if err := importCert(certFile, keyFile); err != nil {
		return err
	}
	if err := importCA(caFile); err != nil {
		return err
	}
	if err := importCRL(crlFile); err != nil {
		return err
	}
	if err := adjustOwner(); err != nil {
		return err
	}
	hwlog.RunLog.Info("[OP]import certificate successfully")
	if notDel {
		hwlog.RunLog.Info("please delete the relevant sensitive files once you decide not to use them again.")
		return nil
	}
	if err := utils.OverridePassWdFile(keyFile, []byte{}, utils.FileMode); err != nil {
		hwlog.RunLog.Warn("security delete key file failed")
	}
	err := os.Remove(keyFile)
	if err != nil {
		hwlog.RunLog.Warn("delete private key file automatically failed,please delete it by yourself")
		return nil
	}
	hwlog.RunLog.Warn("delete private key file automatically")
	return nil
}

func importCert(certFile, keyFile string) error {
	hwlog.RunLog.Info("[OP]start to import the key file")
	keyBlock, err := utils.DecryptPrivateKeyWithPd(keyFile, nil)
	if err != nil {
		return err
	}
	hwlog.RunLog.Info("[OP]start to import the cert file")
	certBytes, err := utils.ReadLimitBytes(certFile, utils.Size10M)
	if err != nil {
		return errors.New("read certFile failed")
	}
	// validate certification and private key, if not pass, program will exit
	if _, err = utils.ValidateX509Pair(certBytes, pem.EncodeToMemory(keyBlock)); err != nil {
		return err
	}
	if err = utils.MakeSureDir(keyStore); err != nil {
		return err
	}
	hwlog.RunLog.Info("encrypt private key again with passwd")
	encryptedBlock, err := utils.EncryptPrivateKeyAgain(keyBlock, passFile, passFileBackUp, encryptAlgorithm)
	if err = utils.OverridePassWdFile(keyStore, pem.EncodeToMemory(encryptedBlock), utils.FileMode); err != nil {
		return err
	}
	hwlog.RunLog.Info("[OP]encrypted key file import successfully")
	if err = ioutil.WriteFile(certStore, certBytes, utils.FileMode); err != nil {
		return errors.New("write certBytes to config failed ")
	}
	hwlog.RunLog.Info("[OP]cert file import successfully")
	return nil
}

func importCA(caFile string) error {
	hwlog.RunLog.Info("[OP]start to import the ca file")
	caBytes, err := utils.CheckCaCert(caFile)
	if err != nil {
		return err
	}
	if len(caBytes) != 0 {
		if err = ioutil.WriteFile(caStore, caBytes, utils.FileMode); err != nil {
			return errors.New("write caBytes to config failed ")
		}
		hwlog.RunLog.Info("[OP]ca file import successfully")
	}
	return nil
}

func importCRL(crlFile string) error {
	// start to import the crl file
	hwlog.RunLog.Info("[OP]start to import the crl file")
	crlBytes, err := utils.CheckCRL(crlFile)
	if err != nil {
		return err
	}
	if len(crlBytes) != 0 {
		if err = ioutil.WriteFile(crlStore, crlBytes, utils.FileMode); err != nil {
			return errors.New("write crlBytes to config failed ")
		}
		hwlog.RunLog.Info("[OP]crl file import successfully")
	}
	return nil
}

func valid(certFile string, keyFile string, caFile string, crlFile string) error {
	if certFile == "" && keyFile == "" && caFile == "" && crlFile == "" {
		return errors.New("no new certificate files need to be imported")
	}
	if certFile == "" || keyFile == "" {
		return errors.New("need input certFile and keyFile together")
	}
	return commonValid()
}

func commonValid() error {
	if encryptAlgorithm != utils.Aes128gcm && encryptAlgorithm != utils.Aes256gcm {
		hwlog.RunLog.Warn("reset invalid encryptAlgorithm ")
		encryptAlgorithm = utils.Aes256gcm
	}
	cp, ok := cptMap[component]
	if !ok {
		return errors.New("the component is invalid")
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
	return checkPathIsExist(paths)
}

func checkPathIsExist(paths []string) error {
	for _, v := range paths {
		if !utils.IsExists(v) {
			continue
		}
		_, err := utils.CheckPath(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func initHwLogger() error {
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
	return hwlog.InitRunLogger(hwLogConfig, context.Background())
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

func importKubeConfig(kubeConf string) error {
	if kubeConf == "" {
		return nil
	}
	hwlog.RunLog.Infof("[OP]start to import kubeConfig and the program version is %s", versions.BuildVersion)
	conf, err := utils.CheckPath(kubeConf)
	if err != nil {
		return err
	}
	if suffix := path.Ext(kubeConf); suffix != ".conf" {
		return errors.New("invalid kubeConfig file")
	}
	if err = commonValid(); err != nil {
		return err
	}
	btes, err := utils.ReadLimitBytes(conf, utils.Size10M)
	if err != nil {
		return err
	}
	if err = utils.KmcInit(encryptAlgorithm, "", ""); err != nil {
		return err
	}
	encryptedConf, err := utils.Encrypt(0, btes)
	if err != nil {
		return errors.New("encrypt kubeConfig failed")
	}
	if err = utils.MakeSureDir(keyStore); err != nil {
		return err
	}
	hwlog.RunLog.Info("[OP]encrypt kubeConfig successfully")
	if err = utils.OverridePassWdFile(kubeConfStore, encryptedConf, utils.FileMode); err != nil {
		return errors.New("write encrypted kubeConfig to file failed")
	}
	hwlog.RunLog.Info("[OP]import kubeConfig successfully")
	if certFile == "" || keyFile == "" {
		return resourceClean(conf)
	}
	return nil
}

func resourceClean(conf string) error {
	if err := adjustOwner(); err != nil {
		return err
	}
	utils.KmcShutDown()
	if notDel {
		hwlog.RunLog.Info("please delete the relevant sensitive files once you decide not to use them again.")
		return nil
	}
	if err := utils.OverridePassWdFile(conf, []byte{}, utils.FileMode); err != nil {
		hwlog.RunLog.Warn("security delete config failed")
	}
	err := os.Remove(conf)
	if err != nil {
		hwlog.RunLog.Warn("delete config file automatically failed,please delete it by yourself")
		return nil
	}
	hwlog.RunLog.Info("delete config file automatically")
	return nil
}

func adjustOwner() error {
	hwlog.RunLog.Info("start to change config file owner")
	filePath, err := utils.CheckPath(dirPrefix)
	if err != nil {
		return errors.New("config file directory is not safe")
	}
	if err := chownR(filePath, hwMindX, hwMindX); err != nil {
		hwlog.RunLog.Warn("change file owner failed, please chown to hwMindX manually")
	}
	hwlog.RunLog.Info("change owner successfully")
	return nil
}
func chownR(path string, uid, gid int) error {
	return filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
		if err == nil {
			err = os.Lchown(name, uid, gid)
		}
		return err
	})
}
