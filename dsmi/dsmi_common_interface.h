//  Copyright(C) 2020. Huawei Technologies Co.,Ltd. All rights reserved.

#ifndef __DSMI_COMMON_INTERFACE_H__
#define __DSMI_COMMON_INTERFACE_H__
#ifdef __cplusplus
extern "C" {
#endif

typedef enum rdfx_detect_result {
    RDFX_DETECT_OK = 0,
    RDFX_DETECT_SOCK_FAIL = 1,
    RDFX_DETECT_RECV_TIMEOUT = 2,
    RDFX_DETECT_UNREACH = 3,
    RDFX_DETECT_TIME_EXCEEDED = 4,
    RDFX_DETECT_FAULT = 5,
    RDFX_DETECT_INIT = 6,
    RDFX_DETECT_MAX
} DSMI_NET_HEALTH_STATUS;

struct dsmi_power_info_stru {
    unsigned short power;
};
struct dsmi_memory_info_stru {
    unsigned long long memory_size;
    unsigned int freq;
    unsigned int utiliza;
};

struct dsmi_hbm_info_stru {
    unsigned long long memory_size;      /**< HBM total size, KB */
    unsigned int freq;          /**< HBM freq, MHZ */
    unsigned long long memory_usage;     /**< HBM memory_usage, KB */
    int temp;                   /**< HBM temperature */
    unsigned int bandwith_util_rate;
};

#define MAX_CHIP_NAME 32
#define MAX_DEVICE_COUNT 64

struct dsmi_chip_info_stru {
    unsigned char chip_type[MAX_CHIP_NAME];
    unsigned char chip_name[MAX_CHIP_NAME];
    unsigned char chip_ver[MAX_CHIP_NAME];
};

#define DSMI_VNIC_PORT 0
#define DSMI_ROCE_PORT 1

enum ip_addr_type {
    IPADDR_TYPE_V4 = 0U,    /**< IPv4 */
    IPADDR_TYPE_V6 = 1U,    /**< IPv6 */
    IPADDR_TYPE_ANY = 2U
};

#define DSMI_ARRAY_IPV4_NUM 4
#define DSMI_ARRAY_IPV6_NUM 16

typedef struct ip_addr {
    union {
        unsigned char ip6[DSMI_ARRAY_IPV6_NUM];
        unsigned char ip4[DSMI_ARRAY_IPV4_NUM];
    } u_addr;
    enum ip_addr_type ip_type;
} ip_addr_t;

/**
* @ingroup driver
* @brief Get the number of devices
* @attention NULL
* @param [out] device_count  The space requested by the user is used to store the number of returned devices
* @return  0 for success, others for fail
* @note Support:Ascend310,Ascend910
*/
int dsmi_get_device_count(int *device_count);

/**
* @ingroup driver
* @brief Get the id of all devices
* @attention NULL
* @param [out] device_id_list[] The space requested by the user is used to store the id of all returned devices
* @param [in] count Number of equipment
* @return  0 for success, others for fail
* @note Support:Ascend310,Ascend910
*/
int dsmi_list_device(int device_id_list[], int count);



/**
* @ingroup driver
* @brief Convert the logical ID of the device to a physical ID
* @attention NULL
* @param [in] logicid logic id
* @param [out] phyid physic id
* @return  0 for success, others for fail
* @note Support:Ascend310,Ascend910
*/
int dsmi_get_phyid_from_logicid(unsigned int logicid, unsigned int *phyid);

/**
* @ingroup driver
* @brief Convert the physical ID of the device to a logical ID
* @attention NULL
* @param [in] phyid   physical id
* @param [out] logicid logic id
* @return  0 for success, others for fail
* @note Support:Ascend310,Ascend910
*/
int dsmi_get_logicid_from_phyid(unsigned int phyid, unsigned int *logicid);

/**
* @ingroup driver
* @brief Query the overall health status of the device, support AI Server
* @attention NULL
* @param [in] device_id  The device id
* @param [out] phealth  The pointer of the overall health status of the device only represents this component,
                        and does not include other components that have a logical relationship with this component.
* @return  0 for success, others for fail
* @note Support:Ascend310,Ascend910
*/
int dsmi_get_device_health(int device_id, unsigned int *phealth);

/**
* @ingroup driver
* @brief Query device fault code
* @attention NULL
* @param [in] device_id The device id.
* @param [out] errorcount Number of error codes, count:0~128
* @param [out] perrorcode error codes
* @return  0 for success, others for fail
* @note Support:Ascend310,Ascend910
*/
int dsmi_get_device_errorcode(int device_id, int *errorcount, unsigned int *perrorcode);

/**
* @ingroup driver
* @brief Query the temperature of the ICE SOC of Ascend AI processor
* @attention NULL
* @param [in] device_id  The device id
* @param [out] ptemperature  The temperature of the HiSilicon SOC of the Shengteng AI processor: unit Celsius,
                         the accuracy is 1 degree Celsius, and the decimal point is rounded. 16-bit signed type,
                         little endian. The value returned by the device is the actual temperature.
* @return  0 for success, others for fail
* @note Support:Ascend310,Ascend910
*/
int dsmi_get_device_temperature(int device_id, int *ptemperature);

/**
* @ingroup driver
* @brief Query device power consumption
* @attention NULL
* @param [in] device_id The device id
* @param [out] pdevice_power_info Device power consumption: unit is W, accuracy is 0.1W. 16-bit unsigned short type,
               little endian
* @return 0 for success, others for fail
* @note Support:Ascend310,Ascend910
*/
int dsmi_get_device_power_info(int device_id, struct dsmi_power_info_stru *pdevice_power_info);


/**
* @ingroup driver
* @brief Query the voltage of Sheng AI SOC of ascend AI processor
* @attention NULL
* @param [in] device_id  The device id
* @param [out] pvoltage  The voltage of the HiSilicon SOC of the Shengteng AI processor: the unit is V,
                         and the accuracy is 0.01V
* @return  0 for success, others for fail
* @note Support:Ascend310,Ascend910
*/
int dsmi_get_device_voltage(int device_id, unsigned int *pvoltage);

/**
* @ingroup driver
* @brief Get the occupancy rate of the HiSilicon SOC of the Ascension AI processor
* @attention NULL
* @param [in] device_id  The device id
* @param [in] device_type  device_type
* @param [out] putilization_rate  Utilization rate of HiSilicon SOC of ascend AI processor, unit:%
* @return  0 for success, others for fail
* @note Support:Ascend310,Ascend910
*/
int dsmi_get_device_utilization_rate(int device_id, int device_type, unsigned int *putilization_rate);

/**
* @ingroup driver
* @brief Get the frequency of the HiSilicon SOC of the Ascension AI processor
* @attention NULL
* @param [in] device_id  The device id
* @param [in] device_type  device_type
* @param [out] pfrequency  Frequency, unit MHZ
* @return  0 for success, others for fail
* @note Support:Ascend310,Ascend910
*/
int dsmi_get_device_frequency(int device_id, int device_type, unsigned int *pfrequency);

/**
* @ingroup driver
* @brief Get memory information
* @attention NULL
* @param [in] device_id  The device id
* @param [out] pdevice_memory_info  Return memory information
* @return  0 for success, others for fail
* @note Support:Ascend310,Ascend910
*/
int dsmi_get_memory_info(int device_id, struct dsmi_memory_info_stru *pdevice_memory_info);


/**
* @ingroup driver
* @brief get the ip address and mask address.
* @attention NULL
* @param [in] device_id  The device id
* @param [in] port_type  Specify the network port type
* @param [in] port_id  Specify the network port number, reserved field
* @param [out] ip_address  return ip address info
* @param [out] mask_address  return mask address info
* @return  0 for success, others for fail
* @note Support:Ascend310,Ascend910
*/
int dsmi_get_device_ip_address(int device_id, int port_type, int port_id, ip_addr_t *ip_address,
    ip_addr_t *mask_address);

/**
* @ingroup driver
* @brief Relevant information about the HiSilicon SOC of the AI ??processor, including chip_type, chip_name,
         chip_ver version number
* @attention NULL
* @param [in] device_id  The device id
* @param [out] chip_info  Get the relevant information of ascend AI processor Hisilicon SOC
* @return  0 for success, others for fail
* @note Support:Ascend310,Ascend910
*/
int dsmi_get_chip_info(int device_id, struct dsmi_chip_info_stru *chip_info);


/**
* @ingroup driver
* @brief Query the frequency, capacity and utilization information of hbm
* @attention NULL
* @param [in] device_id  The device id
* @param [out] pdevice_hbm_info return hbm infomation
* @return  0 for success, others for fail
* @note Support:Ascend910
*/
int dsmi_get_hbm_info(int device_id, struct dsmi_hbm_info_stru *pdevice_hbm_info);


/**
* @ingroup driver
* @brief Query the connectivity status of the RoCE network card's IP address
* @attention NULL
* @param [in] device_id The device id
* @param [out] presult return the result wants to query
* @return  0 for success, others for fail
* @note Support:Ascend910
*/
int dsmi_get_network_health(int device_id, DSMI_NET_HEALTH_STATUS *presult);

#ifdef __cplusplus
}
#endif
#endif
