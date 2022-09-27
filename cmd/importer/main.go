//  Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.

// Package main
package main

import (
	"context"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/tls"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509"
	"huawei.com/npu-exporter/versions"
)

const (
	dirPrefix  = "/etc/mindx-dl/"
	timeFormat = "2006-01-02T15-04-05.000"
	onekilo    = 1000
	hwMindX    = 9000
	// aes128gcm AES128-GCM
	aes128gcm = 8
	// aes256gcm AES256-GCM
	aes256gcm = 9
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
	keyBackup        string
	certStore        string
	certBackup       string
	caStore          string
	caBackup         string
	crlStore         string
	crlBackup        string
	passFile         string
	passFileBackUp   string
	kubeConfig       string
	kubeConfStore    string
	kubeConfBackup   string
	defaultLogFile   = "/var/log/mindx-dl/cert-importer/cert-importer.log"
	cptMap           = map[string]string{
		"ne": "npu-exporter", "am": "access-manager", "lm": "license-manager", "la": "license-agent",
		"hc": "hccl-controller", "dp": "device-plugin", "nd": "noded", "rc": "resilience-controller",
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
	usr, ip, err := utils.GetLoginUserAndIP()
	if err != nil {
		hwlog.RunLog.Warn("get login ip failed")
	}
	hwlog.RunLog.Infof("[OP]current user is %s,hostName is %s,login ip is %s,verison is:%s",
		usr, name, ip, versions.BuildVersion)
	if err = importKubeConfig(kubeConfig); err != nil {
		hwlog.RunLog.Error(err)
		hwlog.RunLog.Error("[OP]kubeConfig imported failed")
		return
	}
	if kubeConfig != "" && (certFile == "" || keyFile == "") {
		hwlog.RunLog.Info(" kubeConfig imported finished")
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
	flag.IntVar(&encryptAlgorithm, "encryptAlgorithm", aes256gcm,
		"Use 8 for aes128gcm,9 for aes256gcm,not recommended config it in general")
	flag.StringVar(&component, "cpt", "ne", "The component name such as ne (npu-exporter),"+
		"am (access-manager),lm(license-manager),la(license agent),hc(hccl-controller),"+
		"dp(device-plugin),nd(noded),rc(resilience-controller)")
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
	if err := importCert(certFile, keyFile); err != nil {
		hwlog.RunLog.Error("[OP] import cert files failed")
		return err
	}
	if err := importCA(caFile); err != nil {
		hwlog.RunLog.Error("[OP] import ca file failed")
		return err
	}
	if err := importCRL(crlFile); err != nil {
		hwlog.RunLog.Error("[OP] import crl file failed")
		return err
	}
	if err := adjustOwner(); err != nil {
		return err
	}
	hwlog.RunLog.Info("import certificate finished")
	if notDel {
		hwlog.RunLog.Info("please delete the relevant sensitive files once you decide not to use them again.")
		return nil
	}
	if err := x509.OverridePassWdFile(keyFile, []byte{}, utils.FileMode); err != nil {
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
	keyBytes, err := utils.ReadLimitBytes(keyFile, utils.Size10M)
	if err != nil {
		hwlog.RunLog.Error("read keyfile failed")
		return err
	}
	keyBlock, err := x509.ParsePrivateKeyWithPassword(keyBytes, nil)
	if err != nil {
		hwlog.RunLog.Error("parsePrivateKeyWithPassword executed error")
		return err
	}
	hwlog.RunLog.Info("[OP]start to import the cert file")
	certBytes, err := utils.ReadLimitBytes(certFile, utils.Size10M)
	if err != nil {
		return errors.New("read certFile failed")
	}
	// validate certification and private key, if not pass, program will exit
	if _, err = tls.ValidateX509Pair(certBytes, pem.EncodeToMemory(keyBlock), x509.InvalidNum); err != nil {
		return err
	}
	if err = utils.MakeSureDir(keyStore); err != nil {
		return err
	}
	hwlog.RunLog.Info("encrypt private key again with passwd")
	encryptedBlock, err := x509.EncryptPrivateKeyAgain(keyBlock, passFile, passFileBackUp, encryptAlgorithm)
	if err != nil {
		return err
	}
	keyBkpInstance, err := x509.NewBKPInstance(pem.EncodeToMemory(encryptedBlock), keyStore, keyBackup)
	if err != nil {
		return err
	}
	if err = keyBkpInstance.WriteToDisk(utils.FileMode, true); err != nil {
		hwlog.RunLog.Error(err)
		return errors.New(" write encrypted key bytes to disk failed ")
	}
	hwlog.RunLog.Info("[OP] key file import successfully")
	certBkpInstance, err := x509.NewBKPInstance(certBytes, certStore, certBackup)
	if err != nil {
		return err
	}
	if err = certBkpInstance.WriteToDisk(utils.FileMode, false); err != nil {
		hwlog.RunLog.Error(err)
		return errors.New(" write certBytes to disk failed ")
	}
	hwlog.RunLog.Info("[OP]cert file import successfully")
	return nil
}

func importCA(caFile string) error {
	if caFile == "" || !utils.IsExist(caFile) {
		return nil
	}
	hwlog.RunLog.Info("[OP]start to import the ca file")
	caBytes, err := x509.CheckCaCert(caFile, x509.InvalidNum)
	if err != nil {
		return err
	}
	if len(caBytes) != 0 {
		bkpInstance, err := x509.NewBKPInstance(caBytes, caStore, caBackup)
		if err != nil {
			return err
		}
		if err = bkpInstance.WriteToDisk(utils.FileMode, false); err != nil {
			hwlog.RunLog.Error(err)
			return errors.New(" write caBytes to disk failed ")
		}
		hwlog.RunLog.Info("[OP]ca file import successfully")
	}
	return nil
}

func importCRL(crlFile string) error {
	if crlFile == "" || !utils.IsExist(crlFile) {
		return nil
	}
	hwlog.RunLog.Info("[OP]start to import the crl file")
	crlBytes, err := x509.CheckCRL(crlFile)
	if err != nil {
		return err
	}
	if len(crlBytes) != 0 {
		bkpInstance, err := x509.NewBKPInstance(crlBytes, crlStore, crlBackup)
		if err != nil {
			return err
		}
		if err = bkpInstance.WriteToDisk(utils.FileMode, false); err != nil {
			hwlog.RunLog.Error(err)
			return errors.New(" write crlBytes to disk failed ")
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
	cp, err := commonValid()
	if err != nil {
		return err
	}
	var paths []string
	keyStore = dirPrefix + cp + "/" + tls.KeyStore
	paths = append(paths, keyStore)
	keyBackup = dirPrefix + cp + "/" + tls.KeyBackup
	paths = append(paths, keyBackup)
	certStore = dirPrefix + cp + "/" + tls.CertStore
	paths = append(paths, certStore)
	certBackup = dirPrefix + cp + "/" + tls.CertBackup
	paths = append(paths, certBackup)
	caStore = dirPrefix + cp + "/" + tls.CaStore
	paths = append(paths, caStore)
	caBackup = dirPrefix + cp + "/" + tls.CaBackup
	paths = append(paths, caBackup)
	crlStore = dirPrefix + cp + "/" + tls.CrlStore
	paths = append(paths, crlStore)
	crlBackup = dirPrefix + cp + "/" + tls.CrlBackup
	paths = append(paths, crlBackup)
	passFile = dirPrefix + cp + "/" + tls.PassFile
	paths = append(paths, passFile)
	passFileBackUp = dirPrefix + cp + "/" + tls.PassFileBackUp
	paths = append(paths, passFileBackUp)
	return checkPathIsExist(paths)
}

func commonValid() (string, error) {
	if encryptAlgorithm != aes128gcm && encryptAlgorithm != aes256gcm {
		hwlog.RunLog.Warn("reset invalid encryptAlgorithm ")
		encryptAlgorithm = aes256gcm
	}
	cp, ok := cptMap[component]
	if !ok {
		return "", errors.New("the component is invalid")
	}
	return cp, nil
}

func kubeValid(kubeConf string) error {
	if suffix := path.Ext(kubeConf); suffix != ".conf" {
		return errors.New("invalid kubeConfig file")
	}
	cp, err := commonValid()
	if err != nil {
		return err
	}
	var paths []string
	kubeConfStore = dirPrefix + cp + "/" + tls.KubeCfgFile
	paths = append(paths, kubeConfStore)
	kubeConfBackup = dirPrefix + cp + "/" + tls.KubeCfgBackup
	paths = append(paths, kubeConfBackup)
	if err = checkPathIsExist(paths); err != nil {
		hwlog.RunLog.Error(err)
		return errors.New("kubeConfig store file check failed")
	}
	return nil
}

func checkPathIsExist(paths []string) error {
	for _, v := range paths {
		_, err := utils.CheckPath(v)
		if err != nil {
			return fmt.Errorf("%s file check failed:%s", utils.MaskPrefix(v), err.Error())
		}
	}
	return nil
}

func initHwLogger() error {
	if utils.IsExist(hwLogConfig.LogFileName) {
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
	hwlog.RunLog.Info("[OP]start to import kubeConfig")
	conf, err := utils.CheckPath(kubeConf)
	if err != nil {
		hwlog.RunLog.Error("check imported path failed")
		return err
	}
	if err = kubeValid(conf); err != nil {
		return err
	}
	configBytes, err := utils.ReadLimitBytes(conf, utils.Size10M)
	if err != nil {
		hwlog.RunLog.Error("read file content failed")
		return err
	}
	if err = kmc.Initialize(encryptAlgorithm, "", ""); err != nil {
		return err
	}
	defer func() {
		if err = kmc.Finalize(); err != nil {
			hwlog.RunLog.Error(err)
		}
	}()
	encryptedConf, err := kmc.Encrypt(0, configBytes)
	if err != nil {
		return errors.New("encrypt kubeConfig failed")
	}
	hwlog.RunLog.Info("encrypt kubeConfig successfully")
	if err = utils.MakeSureDir(kubeConfStore); err != nil {
		return err
	}
	bkpInstance, err := x509.NewBKPInstance(encryptedConf, kubeConfStore, kubeConfBackup)
	if err != nil {
		hwlog.RunLog.Error("new backup instance failed")
		return err
	}
	hwlog.RunLog.Info("start to write data to disk")
	if err = bkpInstance.WriteToDisk(utils.FileMode, true); err != nil {
		hwlog.RunLog.Error(err)
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
	if notDel {
		hwlog.RunLog.Info("please delete the relevant sensitive files once you decide not to use them again.")
		return nil
	}
	if err := x509.OverridePassWdFile(conf, []byte{}, utils.FileMode); err != nil {
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
