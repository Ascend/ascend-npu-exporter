# Description: path config 路径配置
# Note： build在开发/测试中使用，并作为产品编译参考
#        这里没有使用CMAKE_CURRENT_SOURCE_DIR 而使用CMAKE_CURRENT_LIST_DIR

############################# external [编译依赖默认目录，产品需处理必选项1 2 3] #####################
# jdk jni header dir [JNI头文件，需产品提供对应头文件路径，必选项1]
# 提供方式为配置cmake选项 -DJAVA_JNI_INCLUDE_DIR=/path/to/jni/header
# 包含jni.h、jni_md.h，jdk版本要求：jdk1.8*，产品需配置编译环境jdk的jni头文件目录
if(JAVA_JNI_INCLUDE_DIR)
    set(KMC_JAVA_JNI_INCLUDE_DIR
        ${JAVA_JNI_INCLUDE_DIR})
else()
    set(KMC_JAVA_JNI_INCLUDE_DIR
        ${CMAKE_CURRENT_LIST_DIR}/../../external/3rd/jdk/include)
endif()

# hwsec c source dir [华为安全函数库源码目录，需产品提供源码，必选项2]
# 提供方式为配置cmake选项 -DHW_SECUREC_DIR=/path/to/securec
# 华为安全函数库为源码交付，包含src和include目录，产品需配置源码目录
if(HW_SECUREC_DIR)
    set(KMC_EXTERNAL_HWSEC_C_SOURCE_DIR
        ${HW_SECUREC_DIR})
else()
    set(KMC_EXTERNAL_HWSEC_C_SOURCE_DIR 
        ${CMAKE_CURRENT_LIST_DIR}/../../external/huawei/securec/)
endif()

# openssl依赖方式
# 方式一：
# openssl root dir [openssl安装根目录，需产品提供，必选项3]
# 提供方式为配置cmake选项 -DSYSTEM_OPENSSL_PREFIX=/path/to/openssl/prefix
# 方式二：
# openssl dir [openssl压缩包目录，需产品提供openssl压缩包，必选项3]
# 提供方式为配置cmake选项 -DOPENSSL_TARBALL_PATH=/path/to/openssl/tarfile

# 如下为openssl其他配置，产品无需配置
# 如果使用系统自带openssl,且安装有多份openssl版本
# 可通过cmake -DSYSTEM_OPENSSL_PREFIX=/path/to/openssl/prefix 指定openssl根目录
# 例如openssl安装在/usr/local下(包括include lib(lib64)目录)
# cmake -DSYSTEM_OPENSSL_PREFIX=/usr/local/
# 如果使用本地编译需要通过cmake -DUSE_LOCAL_OPENSSL=ON来开启
if(USE_LOCAL_OPENSSL)
    set(KMC_OPENSSL_SOURCE_DIR 
        ${CMAKE_CURRENT_LIST_DIR}/../../external/opensource/openssl)
endif()

############################# source [kmc 源码默认目录，产品无需配置] ################################
# kmc-c source dir
set(KMC_C_SOURCE_DIR
    ${CMAKE_CURRENT_LIST_DIR}/../../src/kmc/)

# kmc ext for java source dir
set(KMCEXT_C_SOURCE_DIR
    ${CMAKE_CURRENT_LIST_DIR}/../../src/kmc-ext)

