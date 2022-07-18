//  Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.

#ifndef __DCMI_INTERFACE_API_H__
#define __DCMI_INTERFACE_API_H__

#ifdef __cplusplus
#if __cplusplus
extern "C" {
#endif
#endif /* __cplusplus */

#define DCMIDLLEXPORT

#define MAX_CHIP_NAME_LEN 32  // Maximum length of chip name
#define TEMPLATE_NAME_LEN 32

/*----------------------------------------------*
 * Structure description                        *
 *----------------------------------------------*/
struct dcmi_chip_info {
    unsigned char chip_type[MAX_CHIP_NAME_LEN];
    unsigned char chip_name[MAX_CHIP_NAME_LEN];
    unsigned char chip_ver[MAX_CHIP_NAME_LEN];
    unsigned int aicore_cnt;
};


struct dcmi_hbm_info {
    unsigned long long memory_size;
    unsigned int freq;
    unsigned long long memory_usage;
    int temp;
    unsigned int bandwith_util_rate;
};

struct dcmi_get_memory_info_stru {
    unsigned long long memory_size;        /* unit:MB */
    unsigned long long memory_available;   /* free + hugepages_free * hugepagesize */
    unsigned int freq;
    unsigned long hugepagesize;             /* unit:KB */
    unsigned long hugepages_total;
    unsigned long hugepages_free;
    unsigned int utiliza;                  /* ddr memory info usages */
    unsigned char reserve[60];             /* the size of dcmi_memory_info is 96 */
};

enum dcmi_ip_addr_type {
    DCMI_IPADDR_TYPE_V4 = 0, /** IPv4 */
    DCMI_IPADDR_TYPE_V6 = 1, /** IPv6 */
    DCMI_IPADDR_TYPE_ANY = 2 /** IPv4+IPv6 ("dual-stack") */
};

struct dcmi_ip_addr {
    union {
        unsigned char ip6[16];
        unsigned char ip4[4];
    } u_addr;
    enum dcmi_ip_addr_type ip_type;
};

enum dcmi_unit_type {
    NPU_TYPE = 0,
    MCU_TYPE = 1,
    CPU_TYPE = 2,
    INVALID_TYPE = 0xFF
};

enum dcmi_rdfx_detect_result {
    DCMI_RDFX_DETECT_OK = 0,
    DCMI_RDFX_DETECT_SOCK_FAIL = 1,
    DCMI_RDFX_DETECT_RECV_TIMEOUT = 2,
    DCMI_RDFX_DETECT_UNREACH = 3,
    DCMI_RDFX_DETECT_TIME_EXCEEDED = 4,
    DCMI_RDFX_DETECT_FAULT = 5,
    DCMI_RDFX_DETECT_INIT = 6,
    DCMI_RDFX_DETECT_THREAD_ERR = 7,
    DCMI_RDFX_DETECT_IP_SET = 8,
    DCMI_RDFX_DETECT_MAX = 0xFF
};

enum dcmi_port_type {
    DCMI_VNIC_PORT = 0,
    DCMI_ROCE_PORT = 1,
    DCMI_INVALID_PORT
};

enum dcmi_main_cmd {
    DCMI_MAIN_CMD_DVPP = 0,
    DCMI_MAIN_CMD_ISP,
    DCMI_MAIN_CMD_TS_GROUP_NUM,
    DCMI_MAIN_CMD_CAN,
    DCMI_MAIN_CMD_UART,
    DCMI_MAIN_CMD_UPGRADE,
    DCMI_MAIN_CMD_TEMP = 50,
    DCMI_MAIN_CMD_SVM = 51,
    DCMI_MAIN_CMD_VDEV_MNG,
    DCMI_MAIN_CMD_DEVICE_SHARE = 0x8001,
    DCMI_MAIN_CMD_MAX
};

enum dcmi_freq_type {
    DCMI_FREQ_DDR = 1,
    DCMI_FREQ_CTRLCPU = 2,
    DCMI_FREQ_HBM = 6,
    DCMI_FREQ_AICORE_CURRENT_ = 7,
    DCMI_FREQ_AICORE_MAX = 9,
    DCMI_FREQ_VECTORCORE_CURRENT = 12
};

#define DCMI_VDEV_RES_NAME_LEN 16
#define DCMI_VDEV_FOR_RESERVE 32
#define DCMI_SOC_SPLIT_MAX 32
struct dcmi_base_resource {
    unsigned long long token;
    unsigned long long token_max;
    unsigned long long task_timeout;
    unsigned int vfg_id;
    unsigned char vip_mode;
    unsigned char reserved[DCMI_VDEV_FOR_RESERVE - 1];  /* bytes aligned */
};

/* total types of computing resource */
struct dcmi_computing_resource {
    /* accelator resource */
    float aic;
    float aiv;
    unsigned short dsa;
    unsigned short rtsq;
    unsigned short acsq;
    unsigned short cdqm;
    unsigned short c_core;
    unsigned short ffts;
    unsigned short sdma;
    unsigned short pcie_dma;

    /* memory resource, MB as unit */
    unsigned long long memory_size;

    /* id resource */
    unsigned int event_id;
    unsigned int notify_id;
    unsigned int stream_id;
    unsigned int model_id;

    /* cpu resource */
    unsigned short topic_schedule_aicpu;
    unsigned short host_ctrl_cpu;
    unsigned short host_aicpu;
    unsigned short device_aicpu;
    unsigned short topic_ctrl_cpu_slot;

    unsigned char reserved[DCMI_VDEV_FOR_RESERVE];
};

struct dcmi_media_resource {
    /* dvpp resource */
    float jpegd;
    float jpege;
    float vpc;
    float vdec;
    float pngd;
    float venc;
    unsigned char reserved[DCMI_VDEV_FOR_RESERVE];
};

struct dcmi_create_vdev_out {
    unsigned int vdev_id;
    unsigned int pcie_bus;
    unsigned int pcie_device;
    unsigned int pcie_func;
    unsigned int vfg_id;
    unsigned char reserved[DCMI_VDEV_FOR_RESERVE];
};

struct dcmi_create_vdev_res_stru {
    unsigned int vdev_id;
    unsigned int vfg_id;
    char template_name[TEMPLATE_NAME_LEN];
    unsigned char reserved[64];
};

struct dcmi_vdev_query_info {
    char name[DCMI_VDEV_RES_NAME_LEN];
    unsigned int status;
    unsigned int is_container_used;
    unsigned int vfid;
    unsigned int vfg_id;
    unsigned long long container_id;
    struct dcmi_base_resource base;
    struct dcmi_computing_resource computing;
    struct dcmi_media_resource media;
};

/* for single search */
struct dcmi_vdev_query_stru {
    unsigned int vdev_id;
    struct dcmi_vdev_query_info query_info;
};

struct dcmi_soc_free_resource {
    unsigned int vfg_num;
    unsigned int vfg_bitmap;
    struct dcmi_base_resource base;
    struct dcmi_computing_resource computing;
    struct dcmi_media_resource media;
};

struct dcmi_soc_total_resource {
    unsigned int vdev_num;
    unsigned int vdev_id[DCMI_SOC_SPLIT_MAX];
    unsigned int vfg_num;
    unsigned int vfg_bitmap;
    struct dcmi_base_resource base;
    struct dcmi_computing_resource computing;
    struct dcmi_media_resource media;
};

#define DCMI_VERSION_1
#define DCMI_VERSION_2

#if defined DCMI_VERSION_2

DCMIDLLEXPORT int dcmi_init(void);

DCMIDLLEXPORT int dcmi_get_card_list(int *card_num, int *card_list, int list_len);

DCMIDLLEXPORT int dcmi_get_device_num_in_card(int card_id, int *device_num);

DCMIDLLEXPORT int dcmi_get_device_id_in_card(int card_id, int *device_id_max, int *mcu_id, int *cpu_id);

DCMIDLLEXPORT int dcmi_get_device_type(int card_id, int device_id, enum dcmi_unit_type *device_type);

DCMIDLLEXPORT int dcmi_get_device_chip_info(int card_id, int device_id, struct dcmi_chip_info *chip_info);

DCMIDLLEXPORT int dcmi_get_device_power_info(int card_id, int device_id, int *power);

DCMIDLLEXPORT int dcmi_get_device_health(int card_id, int device_id, unsigned int *health);

DCMIDLLEXPORT int dcmi_get_device_errorcode_v2(
    int card_id, int device_id, int *error_count, unsigned int *error_code_list, unsigned int list_len);

DCMIDLLEXPORT int dcmi_get_device_temperature(int card_id, int device_id, int *temperature);

DCMIDLLEXPORT int dcmi_get_device_voltage(int card_id, int device_id, unsigned int *voltage);

DCMIDLLEXPORT int dcmi_get_device_frequency(
    int card_id, int device_id, enum dcmi_freq_type input_type, unsigned int *frequency);

DCMIDLLEXPORT int dcmi_get_device_hbm_info(int card_id, int device_id, struct dcmi_hbm_info *hbm_info);

DCMIDLLEXPORT int dcmi_get_device_memory_info_v3(int card_id, int device_id,
    struct dcmi_get_memory_info_stru *memory_info);

DCMIDLLEXPORT int dcmi_get_device_utilization_rate(
    int card_id, int device_id, int input_type, unsigned int *utilization_rate);

DCMIDLLEXPORT int dcmi_get_device_info(
    int card_id, int device_id, enum dcmi_main_cmd main_cmd, unsigned int sub_cmd, void *buf, unsigned int *size);

DCMIDLLEXPORT int dcmi_get_device_ip(int card_id, int device_id, enum dcmi_port_type input_type, int port_id,
    struct dcmi_ip_addr *ip, struct dcmi_ip_addr *mask);

DCMIDLLEXPORT int dcmi_get_device_network_health(int card_id, int device_id, enum dcmi_rdfx_detect_result *result);

DCMIDLLEXPORT int dcmi_get_device_logic_id(int *device_logic_id, int card_id, int device_id);

DCMIDLLEXPORT int dcmi_create_vdevice(int card_id, int device_id, struct dcmi_create_vdev_res_stru *vdev,
    struct dcmi_create_vdev_out *out);

DCMIDLLEXPORT int dcmi_set_destroy_vdevice(int card_id, int device_id, unsigned int vdevid);

DCMIDLLEXPORT int dcmi_get_device_phyid_from_logicid(unsigned int logicid, unsigned int *phyid);

DCMIDLLEXPORT int dcmi_get_device_logicid_from_phyid(unsigned int phyid, unsigned int *logicid);

DCMIDLLEXPORT int dcmi_get_card_id_device_id_from_logicid(int *card_id, int *device_id, unsigned int device_logic_id);

DCMIDLLEXPORT int dcmi_get_card_id_device_id_from_phyid(int *card_id, int *device_id, unsigned int device_phy_id);

#endif

#if defined DCMI_VERSION_1
/* The following interfaces are V1 version interfaces. In order to ensure the compatibility is temporarily reserved,
 * the later version will be deleted. Please switch to the V2 version interface as soon as possible */

struct dcmi_memory_info_stru {
    unsigned long long memory_size;
    unsigned int freq;
    unsigned int utiliza;
};

DCMIDLLEXPORT int dcmi_get_memory_info(int card_id, int device_id, struct dcmi_memory_info_stru *device_memory_info);

DCMIDLLEXPORT int dcmi_get_device_errorcode(
    int card_id, int device_id, int *error_count, unsigned int *error_code, int *error_width);

DCMIDLLEXPORT int dcmi_mcu_get_power_info(int card_id, int *power);
#endif

#ifdef __cplusplus
#if __cplusplus
}
#endif
#endif /* __cplusplus */

#endif /* __DCMI_INTERFACE_API_H__ */
