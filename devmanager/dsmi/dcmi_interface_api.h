/*
* Copyright (c) Huawei Technologies Co., Ltd. 2020-2021. All rights reserved.
* Description: dcmi interface
* Author: huawei
* Create: 2021-03-11
*/

#ifndef _DCMI_INFERENCE_INTERFACE_H_
#define _DCMI_INFERENCE_INTERFACE_H_

#define SENSOR_TEMP_LEN     2
#define SENSOR_NTC_TEMP_LEN 4
#define SENSOR_DATA_MAX_LEN 16
#define DIEID_INFO_LENTH    5
#define BUFF_MAX_LEN        256

int dcmi_init(void);
int dcmi_get_card_num_list(int *card_num, int *card_list, int list_length);
int dcmi_get_device_num_in_card(int card_id, int *device_num);
int dcmi_mcu_get_power_info(int card_id,int *power);
int dcmi_get_device_logic_id(int *device_logic_id, int card_id, int device_id);
#endif /* _DCMI_INFERENCE_INTERFACE_H_ */
